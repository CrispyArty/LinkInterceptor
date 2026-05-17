package system

import (
	"cmp"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

var (
	user32                   = windows.NewLazySystemDLL("user32.dll")
	getForegroundWindow      = user32.NewProc("GetForegroundWindow")
	getWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	getSystemMetrics         = user32.NewProc("GetSystemMetrics")
)

func GetDownloadsDir() (string, error) {
	dir, err := windows.KnownFolderPath(windows.FOLDERID_Downloads, 0)
	if err == nil {
		return dir, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, "Downloads"), nil
}

// func fallbackName(exePath string) string {
// 	return exePath
// }

// func getExeProductName(exePath string) (string, error) {
// 	exePathPtr, err := windows.UTF16PtrFromString(exePath)
// 	if err != nil {
// 		return fallbackName(exePath), nil
// 	}

// 	version := windows.NewLazySystemDLL("version.dll")
// 	getFileVersionInfoSize := version.NewProc("GetFileVersionInfoSizeW")
// 	getFileVersionInfo := version.NewProc("GetFileVersionInfoW")
// 	verQueryValue := version.NewProc("VerQueryValueW")

// 	// get size of version info block
// 	size, _, _ := getFileVersionInfoSize.Call(
// 		uintptr(unsafe.Pointer(exePathPtr)), 0,
// 	)
// 	if size == 0 {
// 		return fallbackName(exePath), nil
// 	}

// 	// read version info block
// 	buf := make([]byte, size)
// 	ret, _, _ := getFileVersionInfo.Call(
// 		uintptr(unsafe.Pointer(exePathPtr)),
// 		0,
// 		size,
// 		uintptr(unsafe.Pointer(&buf[0])),
// 	)
// 	if ret == 0 {
// 		return fallbackName(exePath), nil
// 	}

// 	// query ProductName from version info
// 	// path format: \StringFileInfo\<langID><codepage>\ProductName
// 	subBlock, _ := windows.UTF16PtrFromString(`\StringFileInfo\040904B0\ProductName`)
// 	var productName uintptr
// 	var productNameLen uint32
// 	ret, _, _ = verQueryValue.Call(
// 		uintptr(unsafe.Pointer(&buf[0])),
// 		uintptr(unsafe.Pointer(subBlock)),
// 		uintptr(unsafe.Pointer(&productName)),
// 		uintptr(unsafe.Pointer(&productNameLen)),
// 	)
// 	if ret == 0 || productNameLen == 0 {
// 		return fallbackName(exePath), nil
// 	}

// 	name := windows.UTF16ToString((*[1 << 16]uint16)(unsafe.Pointer(productName))[:productNameLen])
// 	if name == "" {
// 		return fallbackName(exePath), nil
// 	}

// 	return name, nil
// }

type appInfo struct {
	name      string
	icon      string
	shellPath string
	exePath   string
}

func newBrowser(info appInfo) *Browser {
	extractor := ExtractImageFactory(info.icon)
	iconImg, _ := extractor()

	return &Browser{
		Path:      info.exePath,
		Name:      cmp.Or(info.name, filepath.Base(info.exePath)),
		Icon:      info.icon,
		IconImage: iconImg,
		ShellPath: info.shellPath,
	}
}

// Get physical screen resolution
func GetScreenSize() (int, int) {
	// 0 is SM_CXSCREEN (Width), 1 is SM_CYSCREEN (Height)
	w, _, _ := getSystemMetrics.Call(0)
	h, _, _ := getSystemMetrics.Call(1)
	return int(w), int(h)
}

func SplitWindowsArgs(cmdLine string) []string {
	if cmdLine == "" {
		return nil
	}

	// Use Windows API to parse the command line string
	var argc int32
	utf16Ptr, err := syscall.UTF16PtrFromString(cmdLine)
	if err != nil {
		return nil
	}

	// This is the "official" Windows function for splitting command lines
	argvPtr, err := syscall.CommandLineToArgv(utf16Ptr, &argc)
	if err != nil {
		return nil
	}
	// Free the memory allocated by Windows when we are done
	defer syscall.LocalFree(syscall.Handle(unsafe.Pointer(argvPtr)))

	// Convert the pointer array into a Go slice of strings
	var result []string
	// argvPtr is a pointer to an array of pointers to UTF16 strings
	// We create a slice that points to that memory block
	argList := (*[1 << 20]*uint16)(unsafe.Pointer(argvPtr))[:argc:argc]
	for _, arg := range argList {
		result = append(result, syscall.UTF16ToString((*[1 << 20]uint16)(unsafe.Pointer(arg))[:]))
	}

	return result
}

func GetCaller() (string, error) {
	return getCallerProcessName()
}

func GetBrowsers() ([]*Browser, error) {
	roots := []registry.Key{
		registry.CURRENT_USER,
		registry.LOCAL_MACHINE,
	}

	var browsers []*Browser
	seen := make(map[string]bool)

	for _, root := range roots {
		regApps, err := registry.OpenKey(root, `Software\RegisteredApplications`, registry.READ)
		if err != nil {
			continue
		}
		names, err := regApps.ReadValueNames(-1) // gives all names
		defer regApps.Close()

		for _, regName := range names {
			capPath, _, err := regApps.GetStringValue(regName) // Software\Clients\StartMenuInternet\Google Chrome\Capabilities
			if err != nil {
				continue
			}

			urlKey, err := registry.OpenKey(root, capPath+`\URLAssociations`, registry.READ)

			if err != nil {
				continue
			}

			progName, _, err := urlKey.GetStringValue("https") // ChromeHTML

			urlKey.Close()

			if err != nil {
				continue
			}

			if progName == "link-interceptor" {
				continue
			}

			shellPath, err := getExePathFromProgName(progName)

			exePath := parseExePath(shellPath)

			if err != nil || seen[exePath] {
				continue
			}

			seen[exePath] = true

			appInf := getAppInfoFromReg(root, progName, capPath)
			appInf.shellPath = shellPath
			appInf.exePath = exePath

			browsers = append(browsers, newBrowser(appInf))

			// fmt.Printf("proc: %v | %v | %v\n", appInfo.name, exePath, appInfo.icon)
		}
	}

	return browsers, nil
}

// capPath=Software\Clients\StartMenuInternet\Google Chrome\Capabilities
// Example path: HKEY_LOCAL_MACHINE\SOFTWARE\Classes\{regAppName}\Application
// regAppName: ChromeHTML
func getAppInfoFromReg(root registry.Key, regAppName string, capPath string) (app appInfo) {
	paths := []string{
		`SOFTWARE\Classes\` + regAppName + `\Application`,
		capPath,
	}

	for _, path := range paths {

		capKey, err := registry.OpenKey(root, path, registry.READ)

		if err != nil {
			continue
		}

		// log.Println(capPath)
		if appName, _, err := capKey.GetStringValue("ApplicationName"); err == nil && app.name == "" {
			app.name = appName
		}

		if iconName, _, err := capKey.GetStringValue("ApplicationIcon"); err == nil && app.icon == "" {
			app.icon = iconName
		}

		capKey.Close()
	}

	return
}

// func getAppInfoFromCapabilities(root registry.Key, capPath string) (app appInfo) {
// 	capKey, err := registry.OpenKey(root, capPath, registry.READ)

// 	if err != nil {
// 		return
// 	}

// 	log.Println(capPath)
// 	if appName, _, err := capKey.GetStringValue("ApplicationName"); err == nil {
// 		app.name = appName
// 	}

// 	if iconName, _, err := capKey.GetStringValue("ApplicationIcon"); err == nil {
// 		app.icon = iconName
// 	}

// 	capKey.Close()

// 	return
// }

func getExePathFromProgName(name string) (string, error) {
	roots := []registry.Key{
		registry.CURRENT_USER,
		registry.LOCAL_MACHINE,
		registry.CLASSES_ROOT,
	}

	for _, root := range roots {
		// if root == registry.CLASSES_ROOT {
		// 	panic(fmt.Sprintf("!!!!!! search in registry.CLASSES_ROOT for %v", name))
		// }

		shellKey, err := registry.OpenKey(root, `Software\Classes\`+name+`\shell\open\command`, registry.READ)

		if err != nil {
			continue
		}

		shellPath, _, err := shellKey.GetStringValue("")

		shellKey.Close()

		if err != nil {
			continue
		}

		// app := application{path: parseExePath(shellPath)}
		// app.name = filepath.Base(app.path)

		// appKey, err := registry.OpenKey(root, `Software\Classes\`+name+`\Application`, registry.READ)

		// fmt.Printf("name %v | %v\n", name, err)
		// if err == nil {
		// 	if appName, _, err := appKey.GetStringValue("ApplicationName"); err != nil {
		// 		app.name = appName
		// 	}

		// 	if iconName, _, err := appKey.GetStringValue("ApplicationIcon"); err != nil {
		// 		app.icon = iconName
		// 	}
		// }

		return shellPath, nil
	}

	return "", fmt.Errorf("not found shell executable for \"%v\"", name)
}

// parses exe path from shell\open\command
// Examples:
// "C:\Program Files\Firefox\firefox.exe" -osint -url "%1"
// C:\ProgramFiles\Chrome\chrome.exe --single-argument %1
func parseExePath(shellPath string) string {
	str := strings.TrimSpace(shellPath)

	if strings.HasPrefix(str, `"`) {
		end := strings.Index(str[1:], `"`)
		if end >= 0 {
			return str[1 : end+1]
		}
	}

	parts := strings.SplitN(str, " ", 2)
	return filepath.Clean(parts[0])
}

func getCallerProcessName() (string, error) {
	// Get the foreground window handle
	hwnd, _, _ := getForegroundWindow.Call()
	if hwnd == 0 {
		return "", fmt.Errorf("no foreground window")
	}

	// Get the process ID from the window handle
	var pid uint32
	getWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))

	// Open the process and get its executable path
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return "", fmt.Errorf("open proc error: %w", err)
	}
	defer windows.CloseHandle(handle)

	buf := make([]uint16, 260)
	size := uint32(len(buf))
	err = windows.QueryFullProcessImageName(handle, 0, &buf[0], &size)
	if err != nil {
		return "", err
	}

	return windows.UTF16ToString(buf[:size]), nil
}

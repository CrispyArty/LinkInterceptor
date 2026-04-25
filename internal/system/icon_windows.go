package system

import (
	"cmp"
	"fmt"
	"image"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"image/draw"
	_ "image/png"
	"os"
)

var (
	moduser32 = syscall.NewLazyDLL("user32.dll")
	modgdi32  = syscall.NewLazyDLL("gdi32.dll")

	procPrivateExtractIconsW = moduser32.NewProc("PrivateExtractIconsW")
	procGetIconInfo          = moduser32.NewProc("GetIconInfo")
	procGetDC                = moduser32.NewProc("GetDC")
	procReleaseDC            = moduser32.NewProc("ReleaseDC")
	procDestroyIcon          = moduser32.NewProc("DestroyIcon")
	procGetObjectW           = modgdi32.NewProc("GetObjectW")
	procGetDIBits            = modgdi32.NewProc("GetDIBits")
	procDeleteObject         = modgdi32.NewProc("DeleteObject")
)

type ICONINFO struct {
	FIcon    int32
	XHotspot uint32
	YHotspot uint32
	HbmMask  syscall.Handle
	HbmColor syscall.Handle
}

type BITMAP struct {
	Type       int32
	Width      int32
	Height     int32
	WidthBytes int32
	Planes     uint16
	BitsPixel  uint16
	Bits       uintptr
}

type BITMAPINFOHEADER struct {
	BiSize          uint32
	BiWidth         int32
	BiHeight        int32
	BiPlanes        uint16
	BiBitCount      uint16
	BiCompression   uint32
	BiSizeImage     uint32
	BiXPelsPerMeter int32
	BiYPelsPerMeter int32
	BiClrUsed       uint32
	BiClrImportant  uint32
}

// ExtractIconToRGBA extracts the main icon from an .exe file to an image.RGBA
func ExtractIconToRGBA(exePath string, iconIndex, size uint) (*image.RGBA, error) {
	pathPtr, err := syscall.UTF16PtrFromString(exePath)
	if err != nil {
		return nil, err
	}

	var hIcon syscall.Handle
	var iconId uint32
	// 1. Extract the icon (Requesting 256x256, OS will scale if not available)
	ret, _, _ := procPrivateExtractIconsW.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(iconIndex),        // Extract the first icon (index 0)
		cmp.Or(uintptr(size), 64), // Requested Width
		cmp.Or(uintptr(size), 64), // Requested Height
		uintptr(unsafe.Pointer(&hIcon)),
		uintptr(unsafe.Pointer(&iconId)),
		1, // Number of icons to extract
		0, // Flags
	)

	if ret == 0 || hIcon == 0 {
		return nil, fmt.Errorf("failed to extract icon from %s", exePath)
	}
	defer procDestroyIcon.Call(uintptr(hIcon)) // Prevent GDI leaks

	// 2. Get Icon Info (Provides Color and Mask Bitmaps)
	var iconInfo ICONINFO
	ret, _, _ = procGetIconInfo.Call(uintptr(hIcon), uintptr(unsafe.Pointer(&iconInfo)))
	if ret == 0 {
		return nil, fmt.Errorf("GetIconInfo failed")
	}
	defer procDeleteObject.Call(uintptr(iconInfo.HbmMask))
	defer procDeleteObject.Call(uintptr(iconInfo.HbmColor))

	// 3. Get width and height from the Color Bitmap
	var bmp BITMAP
	ret, _, _ = procGetObjectW.Call(
		uintptr(iconInfo.HbmColor),
		uintptr(unsafe.Sizeof(bmp)),
		uintptr(unsafe.Pointer(&bmp)),
	)
	if ret == 0 {
		return nil, fmt.Errorf("GetObjectW failed")
	}

	width := int(bmp.Width)
	height := int(bmp.Height)

	// 4. Prepare Device Context and BitmapHeader to read the pixels
	hdc, _, _ := procGetDC.Call(0)
	defer procReleaseDC.Call(0, hdc)

	bmi := BITMAPINFOHEADER{
		BiSize:        uint32(unsafe.Sizeof(BITMAPINFOHEADER{})),
		BiWidth:       int32(width),
		BiHeight:      int32(-height), // Negative height means top-down drawing
		BiPlanes:      1,
		BiBitCount:    32, // Force 32-bit ARGB
		BiCompression: 0,  // BI_RGB (uncompressed)
	}

	pixels := make([]byte, width*height*4)

	// 5. Read the raw pixels from the OS
	ret, _, _ = procGetDIBits.Call(
		hdc,
		uintptr(iconInfo.HbmColor),
		0,
		uintptr(height),
		uintptr(unsafe.Pointer(&pixels[0])),
		uintptr(unsafe.Pointer(&bmi)),
		0, // DIB_RGB_COLORS
	)
	if ret == 0 {
		return nil, fmt.Errorf("GetDIBits failed")
	}

	// 6. Windows stores as BGRA, but Go image.RGBA needs RGBA.
	// Older icons might also not have an alpha channel mapped (all 0s).
	hasAlpha := false
	for i := 3; i < len(pixels); i += 4 {
		if pixels[i] > 0 {
			hasAlpha = true
			break
		}
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for i := 0; i < len(pixels); i += 4 {
		img.Pix[i] = pixels[i+2]   // Red   (from Blue)
		img.Pix[i+1] = pixels[i+1] // Green
		img.Pix[i+2] = pixels[i]   // Blue  (from Red)

		if hasAlpha {
			img.Pix[i+3] = pixels[i+3] // Alpha
		} else {
			img.Pix[i+3] = 255 // Force opaque if 24-bit legacy icon
		}
	}

	return img, nil
}

var (
	modshlwapi               = syscall.NewLazyDLL("shlwapi.dll")
	procSHLoadIndirectString = modshlwapi.NewProc("SHLoadIndirectString")
)

func LoadIndirectString(source string) (string, error) {
	outBuf := make([]uint16, 1024)
	str, _ := syscall.UTF16PtrFromString(source)
	ret, _, _ := procSHLoadIndirectString.Call(
		uintptr(unsafe.Pointer(str)),
		uintptr(unsafe.Pointer(&outBuf[0])),
		uintptr(len(outBuf)),
		0,
	)
	if ret != 0 {
		return "", fmt.Errorf("failed to load string")
	}
	return syscall.UTF16ToString(outBuf), nil
}

func loadPngAsRGBA(filePath string) (*image.RGBA, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	if rgba, ok := img.(*image.RGBA); ok {
		return rgba, nil
	}

	bounds := img.Bounds()
	rgba := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(rgba, rgba.Bounds(), img, bounds.Min, draw.Src)

	return rgba, nil
}

// Examples:
// @{TheBrowserCompany.Arc_1.101.0.289_x64__ttt1ap7aakyb4?ms-resource://Application/Files/assets/AppList.png}
// C:\Program Files\Google\Chrome\Application\chrome.exe,0
func ExtractImageFactory(iconPath string) func() (*image.RGBA, error) {
	if strings.Contains(iconPath, `ms-resource://`) {
		return func() (*image.RGBA, error) {
			path, err := LoadIndirectString(iconPath)
			if err != nil {
				return nil, err
			}

			return loadPngAsRGBA(path)
		}
	} else {
		return func() (*image.RGBA, error) {
			parts := strings.SplitN(iconPath, ",", 2)
			path := parts[0]
			var index int
			if len(parts) > 1 {
				index, _ = strconv.Atoi(parts[1])
			}

			return ExtractIconToRGBA(path, uint(index), 64)
		}
	}
}

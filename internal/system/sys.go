package system

import (
	"fmt"
	"image"

	"gioui.org/op/paint"
	"github.com/sqweek/dialog"
)

type Browser struct {
	Name        string
	Icon        string
	IconImage   *image.RGBA
	IconImageOp *paint.ImageOp
	Path        string
	ShellPath   string
}

func OpenDialog(startFilename string) (string, error) {
	startDir, _ := GetDownloadsDir()

	path, err := dialog.File().
		SetStartDir(startDir).
		SetStartFile(startFilename).
		// Filter("Torrent file", "torrent").
		Title("Save to").
		Save()
	if err != nil {
		// fmt.Println("cancelled")
		return "", fmt.Errorf("cancelled")
	}

	return path, nil
}

// func OpenPath(path string) error {
// 	var cmd *exec.Cmd

// 	switch runtime.GOOS {
// 	case "windows":
// 		cmd = exec.Command("explorer", path)
// 	case "darwin": // macOS
// 		cmd = exec.Command("open", path)
// 	case "linux":
// 		cmd = exec.Command("xdg-open", path)
// 	default:
// 		return fmt.Errorf("unsupported platform")
// 	}

// 	return cmd.Run()
// }

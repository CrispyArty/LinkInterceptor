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

package system

import (
	"fmt"
	"image"
	"os"

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
	home, _ := os.UserHomeDir()

	path, err := dialog.File().
		SetStartDir(home + "\\Downloads").
		SetStartFile(startFilename).
		// Filter("Torrent file", "torrent").
		Title("Save to").
		Save()
	if err != nil {
		// fmt.Println("cancelled")
		return "", fmt.Errorf("cancelled")
	}
	fmt.Println("Save to:", path)
	return path, nil
}

package uicore

import (
	"image/color"
)

var (
	colorBorderGray = color.NRGBA{218, 220, 224, 255}
	// colorBorderGray = color.NRGBA{0, 220, 0, 255}
	colorBgLightGray = color.NRGBA{235, 237, 240, 255}
	// colorBgLightGray = color.NRGBA{235, 0, 0, 255}
)

// 218, 220, 224

// --backgroundLightGray: rgb(235, 237, 240)

type ColorPalette struct {
	Border, BackgroundLightGray, BackgroundWhite color.NRGBA
}

var Colors = &ColorPalette{
	Border:              colorBorderGray,
	BackgroundLightGray: colorBgLightGray,
	BackgroundWhite:     color.NRGBA{255, 255, 255, 255},
}

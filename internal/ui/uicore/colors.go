package uicore

import (
	"image/color"
	"strconv"
)

var (
	colorBorderGray = HexToColor("#dadce0")
	// colorBorderGray = color.NRGBA{0, 220, 0, 255}
	colorBgLightGray = color.NRGBA{235, 237, 240, 255}
	colorBgGray      = HexToColor("#cccccc")
	// colorBgLightGray = color.NRGBA{235, 0, 0, 255}

	colorBgDarkGray = HexToColor("#6c757d")
)

// 218, 220, 224

// --backgroundLightGray: rgb(235, 237, 240)

type ColorPalette struct {
	Border, BackgroundLightGray, BackgroundWhite, ButtonSecondaryBg, Black, ButtonIconBg color.NRGBA
}

var Colors = &ColorPalette{
	Border:              colorBorderGray,
	BackgroundLightGray: colorBgLightGray,
	BackgroundWhite:     color.NRGBA{255, 255, 255, 255},
	ButtonSecondaryBg:   colorBgDarkGray,
	Black:               color.NRGBA{A: 255},
	ButtonIconBg:        colorBgGray,
}

// Hex translates a string like "5c636a" into color.NRGBA
func HexToColor(hexStr string) color.NRGBA {
	// Remove '#' if present
	if len(hexStr) > 0 && hexStr[0] == '#' {
		hexStr = hexStr[1:]
	}

	// Parse the hex string to a uint64
	hex, err := strconv.ParseUint(hexStr, 16, 64)
	if err != nil {
		return color.NRGBA{A: 255} // Fallback to black
	}

	// If it's a 6-digit hex (RRGGBB)
	if len(hexStr) == 6 {
		return color.NRGBA{
			R: uint8(hex >> 16),
			G: uint8(hex >> 8 & 0xff),
			B: uint8(hex & 0xff),
			A: 255,
		}
	}

	// If it's an 8-digit hex (RRGGBBAA)
	return color.NRGBA{
		R: uint8(hex >> 24),
		G: uint8(hex >> 16 & 0xff),
		B: uint8(hex >> 8 & 0xff),
		A: uint8(hex & 0xff),
	}
}

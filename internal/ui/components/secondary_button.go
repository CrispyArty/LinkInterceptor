package components

import (
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/crispyarty/LinkInterceptor/internal/ui/uicore"
)

func SecondaryButton(th *material.Theme, click *widget.Clickable, txt string) material.ButtonStyle {
	btn := material.Button(th, click, txt)

	btn.Background = uicore.Colors.ButtonSecondaryBg
	// btn.Color = color.NRGBA{A: 255}

	return btn
}

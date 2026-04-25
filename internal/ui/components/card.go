package components

import (
	"image"
	"image/color"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/crispyarty/LinkInterceptor/internal/ui/uicore"
)

type Card struct{}

func (t *Card) Layout(gtx layout.Context, btn *widget.Clickable, w layout.Widget) layout.Dimensions {
	return btn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		pointer.CursorPointer.Add(gtx.Ops)

		var bg color.NRGBA
		if btn.Hovered() {
			bg = uicore.Colors.BackgroundLightGray
		} else {
			bg = uicore.Colors.BackgroundWhite
		}

		borderWidth := unit.Dp(1)
		radius := unit.Dp(4)

		return layout.Stack{}.Layout(gtx,
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				rect := image.Rectangle{Max: gtx.Constraints.Min}.Inset(gtx.Dp(borderWidth))
				radius := gtx.Dp(radius)

				paint.FillShape(gtx.Ops, bg, clip.RRect{
					Rect: rect,
					NW:   radius, NE: radius, SE: radius, SW: radius,
				}.Op(gtx.Ops))

				return layout.Dimensions{Size: gtx.Constraints.Min}
			}),

			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X = gtx.Constraints.Max.X

				return widget.Border{
					Color:        uicore.Colors.Border,
					CornerRadius: radius,
					Width:        borderWidth,
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.UniformInset(unit.Dp(8)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return w(gtx)
					})
				})
			}),
		)
	})
}

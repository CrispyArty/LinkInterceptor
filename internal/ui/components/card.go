package components

import (
	"image"
	"image/color"

	// "gioui.org/internal/f32"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/crispyarty/LinkInterceptor/internal/ui/uicore"
)

type CardButton struct{}

func (t *CardButton) Layout(gtx layout.Context, btn *widget.Clickable, w layout.Widget) layout.Dimensions {
	// fmt.Println("CardButton2 - gtx.Constraints.Max.Y", gtx.Constraints.Max.Y)
	// bg := uicore.Colors.BackgroundLightGray
	// return testCard(gtx, bg, w)

	return btn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		pointer.CursorPointer.Add(gtx.Ops)

		var bg color.NRGBA
		if btn.Hovered() {
			bg = uicore.Colors.BackgroundLightGray
		} else {
			bg = uicore.Colors.BackgroundWhite
		}

		card := DefaultCard
		card.bgColor = bg

		return card.Layout(gtx, w)
	})
}

type Card struct {
	radius      unit.Dp
	borderWidth unit.Dp
	padding     unit.Dp
	borderColor color.NRGBA
	bgColor     color.NRGBA
}

var DefaultCard = Card{
	radius:      unit.Dp(4),
	borderWidth: unit.Dp(2),
	padding:     unit.Dp(8),
	bgColor:     uicore.Colors.BackgroundWhite,
	borderColor: uicore.Colors.Border,
}

func (t Card) Layout(gtx layout.Context, w layout.Widget) layout.Dimensions {
	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			bw := gtx.Dp(t.borderWidth)
			halfBW := bw / 2

			r := gtx.Dp(t.radius)
			// r := gtx.Dp(radius) - halfBW
			rect := image.Rectangle{Max: gtx.Constraints.Min}.Inset(halfBW)

			rr := clip.RRect{
				Rect: rect,
				NW:   r, NE: r, SE: r, SW: r,
			}

			// Background
			paint.FillShape(gtx.Ops, t.bgColor, rr.Op(gtx.Ops))

			// Border
			paint.FillShape(gtx.Ops, t.borderColor, clip.Stroke{
				Path:  rr.Path(gtx.Ops),
				Width: float32(bw),
			}.Op())

			return layout.Dimensions{Size: gtx.Constraints.Min}
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X

			fullInset := t.padding + t.borderWidth

			return layout.UniformInset(fullInset).Layout(gtx, w)
		}),
	)
}

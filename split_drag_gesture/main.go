package main

import (
	"image"
	"image/color"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/gesture"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

func main() {
	go func() {
		// create new window
		w := new(app.Window)

		if err := run(w); err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}()

	app.Main()
}

func run(window *app.Window) (err error) {
	window.Option(app.Title("Link Intercept"))
	window.Option(app.Size(unit.Dp(400), unit.Dp(600)))

	var ops op.Ops

	theme := material.NewTheme()

	for {
		// first grab the event
		evt := window.Event()
		// fmt.Printf("event: %T\n", evt)

		switch typ := evt.(type) {
		case app.FrameEvent:
			gtx := app.NewContext(&ops, typ)
			test(gtx, theme)
			typ.Frame(gtx.Ops)
		case app.DestroyEvent:
			return err
		}
	}

}

func Card(gtx layout.Context, th *material.Theme, bg color.NRGBA, text string) layout.Dimensions {
	width := gtx.Constraints.Max.X
	height := gtx.Constraints.Max.Y

	rect := clip.Rect{
		Max: image.Point{X: width, Y: height},
	}

	paint.FillShape(gtx.Ops, bg, rect.Op())

	layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		lbl := material.Label(th, unit.Sp(24), text)
		lbl.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		return lbl.Layout(gtx)
	})

	return layout.Dimensions{
		Size: image.Point{X: width, Y: height},
	}
}

var split Split

func test(gtx layout.Context, theme *material.Theme) {
	split.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return Card(gtx, theme, color.NRGBA{R: 0, G: 150, B: 0, A: 255}, "Left")
	}, func(gtx layout.Context) layout.Dimensions {
		return Card(gtx, theme, color.NRGBA{R: 0, G: 0, B: 150, A: 255}, "Right")
	})
}

type Split struct {
	drag gesture.Drag

	// Ratio keeps the current layout.
	// 0 is center, -1 completely to the left, 1 completely to the right.
	Ratio float32

	Bar unit.Dp

	// drag   bool
	// dragID pointer.ID
	// dragX  float32
}

const defaultBarWidth = unit.Dp(50)

func (s *Split) Layout(gtx layout.Context, left, right layout.Widget) layout.Dimensions {
	bar := gtx.Dp(s.Bar)
	if bar <= 1 {
		bar = gtx.Dp(defaultBarWidth)
	}

	proportion := (s.Ratio + 1) / 2
	x := float32(gtx.Constraints.Max.X)

	leftsize := int(proportion*x - float32(bar))

	rightoffset := leftsize + bar
	rightsize := gtx.Constraints.Max.X - rightoffset
	// log.Printf("Layout\n")

	{ // handle input
		barRect := image.Rect(leftsize, 0, rightoffset, gtx.Constraints.Max.Y)
		area := clip.Rect(barRect).Push(gtx.Ops)

		// register for input
		// event.Op(gtx.Ops, s)
		pointer.CursorColResize.Add(gtx.Ops)

		log.Printf("##############DRAW##############")

		for {
			ev, ok := s.drag.Update(gtx.Metric, gtx.Source, gesture.Horizontal)

			if !ok {
				break
			}

			log.Printf("s.Ratio %v\n", ev.Position.X)

			deltaRatio := (ev.Position.X*2 - x) / x
			log.Printf("s.Ratio %v\n", s.Ratio)
			s.Ratio = deltaRatio
		}
		s.drag.Add(gtx.Ops)

		area.Pop()
	}
	{
		gtx := gtx
		gtx.Constraints = layout.Exact(image.Pt(leftsize, gtx.Constraints.Max.Y))
		left(gtx)
	}

	{
		trans := op.Offset(image.Pt(rightoffset, 0)).Push(gtx.Ops)
		gtx := gtx
		gtx.Constraints = layout.Exact(image.Pt(rightsize, gtx.Constraints.Max.Y))
		right(gtx)
		trans.Pop()
	}

	return layout.Dimensions{Size: gtx.Constraints.Max}
}

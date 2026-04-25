package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"os"

	"github.com/sqweek/dialog"

	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

func openDialog() {

	home, _ := os.UserHomeDir()

	path, err := dialog.File().
		SetStartDir(home+"\\Downloads").
		SetStartFile("myfile.txt").
		Filter("Torrent file", "torrent").
		Title("Export to XML").
		Save()
	if err != nil {
		fmt.Println("cancelled")
		return
	}
	fmt.Println("Save to:", path)
}

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

	// layout.Center.Layout

	layout.Exact(image.Pt(12, 12))

	var ops op.Ops
	// var startButton widget.Clickable
	// var startButton2 widget.Clickable

	theme := material.NewTheme()

	for {
		// first grab the event
		evt := window.Event()

		// fmt.Printf("event: %T\n", evt)

		switch typ := evt.(type) {
		case app.FrameEvent:
			gtx := app.NewContext(&ops, typ)
			test(gtx, theme)

			// if startButton.Clicked(gtx) {
			// 	fmt.Println("Button1 clicked1!")
			// }

			// if startButton2.Clicked(gtx) {
			// 	fmt.Println("Button2 clicked1!")
			// }

			// title := material.H1(theme, "Gosu1, Gio")
			// title.Layout(gtx)
			// title2 := material.H1(theme, "Hello, Gio2")
			// title2.Layout(gtx)

			// btn := material.Button(theme, &startButton, "Start")
			// btn.Layout(gtx)

			// btn2 := material.Button(theme, &startButton2, "End2")
			// btn2.Layout(gtx)

			typ.Frame(gtx.Ops)

		case app.DestroyEvent:
			return err
		}
	}

}

func Card(gtx layout.Context, th *material.Theme, bg color.NRGBA, text string) layout.Dimensions {
	width := gtx.Constraints.Max.X
	height := gtx.Constraints.Max.Y

	// rrect := clip.RRect{
	// 	Rect: image.Rectangle{
	// 		Max: image.Point{X: width, Y: height},
	// 	},
	// }
	// paint.FillShape(gtx.Ops, bg, rrect.Op(gtx.Ops))

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
	// Ratio keeps the current layout.
	// 0 is center, -1 completely to the left, 1 completely to the right.
	Ratio float32

	Bar unit.Dp

	drag   bool
	dragID pointer.ID
	dragX  float32
}

const defaultBarWidth = unit.Dp(50)

func (s *Split) Layout(gtx layout.Context, left, right layout.Widget) layout.Dimensions {
	bar := gtx.Dp(s.Bar)
	if bar <= 1 {
		bar = gtx.Dp(defaultBarWidth)
	}

	proportion := (s.Ratio + 1) / 2
	leftsize := int(proportion*float32(gtx.Constraints.Max.X) - float32(bar))

	rightoffset := leftsize + bar
	rightsize := gtx.Constraints.Max.X - rightoffset
	// log.Printf("Layout\n")

	{ // handle input
		barRect := image.Rect(leftsize, 0, rightoffset, gtx.Constraints.Max.Y)
		area := clip.Rect(barRect).Push(gtx.Ops)

		gtx.Focused(s)
		// register for input
		event.Op(gtx.Ops, s)
		pointer.CursorColResize.Add(gtx.Ops)

		log.Printf("!!!!!!!!DRAW!!!!!!!!!!!!!")

		for {
			ev, ok := gtx.Event(pointer.Filter{
				Target: s,
				Kinds:  pointer.Press | pointer.Drag | pointer.Release | pointer.Cancel | pointer.Move,
			})
			if !ok {
				break
			}

			e, ok := ev.(pointer.Event)
			if !ok {
				continue
			}

			log.Printf("Event: %T | %v\n", e.Kind, e.Kind)

			switch e.Kind {
			case pointer.Press:
				log.Printf("pointer.Press: %v\n", s.drag)
				if s.drag {
					break
				}

				s.dragID = e.PointerID
				s.dragX = e.Position.X
				s.drag = true

			case pointer.Drag:
				if s.dragID != e.PointerID {
					break
				}

				deltaX := e.Position.X - s.dragX
				s.dragX = e.Position.X

				deltaRatio := deltaX * 2 / float32(gtx.Constraints.Max.X)
				log.Printf("s.Ratio %v\n", s.Ratio)
				s.Ratio += deltaRatio

				if e.Priority < pointer.Grabbed {
					gtx.Execute(pointer.GrabCmd{
						Tag: s,
						ID:  s.dragID,
					})
				}

			case pointer.Release:
				log.Printf("FOCUS!!\n")
				gtx.Execute(key.FocusCmd{Tag: s}) // ← request focus
				fallthrough
			case pointer.Cancel:
				s.drag = false
			}
		}
		if gtx.Focused(s) {
			log.Printf("FOCUSED!!!\n")
		}

		for {
			ev, ok := gtx.Event(
				key.FocusFilter{Target: s}, // Retaines Focus. There is a check at the end of Router.Frame that clears the focus if the tag turns out to fail the requirements (visible and has asked for FocusEvents).
				key.Filter{
					Focus: s,
					Name:  "C",
				})
			if !ok {
				break
			}

			e, ok := ev.(key.Event)

			if !ok {
				focus, ok := ev.(key.FocusEvent)

				if ok {
					log.Printf("~~~~FOCUSED: %v\n", focus.Focus)
				}

				continue
			}

			if e.State == key.Press {
				log.Println("~~~~C pressed~~~~~")
			}
		}

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

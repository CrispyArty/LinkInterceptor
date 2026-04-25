package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

func main() {
	go func() {
		// create new window
		w := new(app.Window)
		w.Option(app.Title("Egg timer"))
		w.Option(app.Size(unit.Dp(400), unit.Dp(600)))

		// listen for events in the window

		if err := run(w); err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}()
	app.Main()
}

// var progress float32
var progressIncrementer chan float32

// var boiling bool
// var boilDuration float32
var boilDurationInput widget.Editor

type animation struct {
	start    time.Time
	duration time.Duration
}

var anim animation

func (a *animation) animate(gtx layout.Context, duration time.Duration) {
	a.start = gtx.Now
	a.duration = duration
	cmd := op.InvalidateCmd{}
	gtx.Execute(cmd)
}

// stop ends the animation immediately.
func (a *animation) stop() {
	a.duration = time.Duration(0)
}

// progress returns whether the animation is currently running and (if so) how far through the animation it is.
func (a animation) progress(asd layout.Context) (animating bool, progress float32) {
	if asd.Now.After(a.start.Add(a.duration)) {
		return false, 0
	}
	cmd := op.InvalidateCmd{}
	asd.Execute(cmd)
	return true, float32(asd.Now.Sub(a.start)) / float32(a.duration)
}

func handleClick(gtx layout.Context) {
	boiling, _ := anim.progress(gtx)

	if boiling {
		anim.stop()
	} else {
		// Read from the input box
		inputString := boilDurationInput.Text()
		inputString = strings.TrimSpace(inputString)
		inputFloat, _ := strconv.ParseFloat(inputString, 32)
		anim.animate(gtx, time.Duration(inputFloat)*time.Second)
	}

	// if progress >= 1 {
	// 	progress = 0
	// }

	// boiling = !boiling

	// inputString := boilDurationInput.Text()
	// inputString = strings.TrimSpace(inputString)
	// inputFloat, _ := strconv.ParseFloat(inputString, 32)

	// // inputFloat = 10
	// boilDuration = float32(inputFloat)
	// // boilDuration = 10

	// boilDuration = boilDuration / (1 - progress)
}

func run(w *app.Window) (err error) {
	var ops op.Ops
	var startButton widget.Clickable

	th := material.NewTheme()

	progressIncrementer = make(chan float32)
	// go func() {
	// 	for {
	// 		time.Sleep(time.Second / 25)
	// 		progressIncrementer <- 0.004
	// 	}
	// }()

	// go func() {
	// 	for range progressIncrementer {
	// 		if boiling && progress < 1 {
	// 			progress += 1.0 / 25.0 / boilDuration
	// 			if progress >= 1 {
	// 				progress = 1
	// 			}
	// 			// w.Invalidate()
	// 		}
	// 	}
	// }()

	for {

		switch evt := w.Event().(type) {
		case app.FrameEvent:
			gtx := app.NewContext(&ops, evt)
			boiling, progress := anim.progress(gtx)

			if startButton.Clicked(gtx) {
				handleClick(gtx)
			}

			layout.Flex{
				// Vertical alignment, from top to bottom
				Axis: layout.Vertical,
				// Empty space is left at the start, i.e. at the top
				Spacing: layout.SpaceStart,
			}.Layout(gtx,
				layout.Rigid(
					func(gtx layout.Context) layout.Dimensions {

						var eggPath clip.Path
						op.Offset(image.Pt(gtx.Dp(200), gtx.Dp(125))).Add(gtx.Ops)
						eggPath.Begin(gtx.Ops)

						for deg := 0.0; deg <= 360; deg++ {
							rad := deg * math.Pi / 180
							cosT := math.Cos(rad)
							sinT := math.Sin(rad)
							a := 110.0
							b := 150.0
							d := 20.0
							x := a * cosT
							y := -(math.Sqrt(b*b-d*d*cosT*cosT) + d*sinT) * sinT
							p := f32.Pt(float32(x), float32(y))
							eggPath.LineTo(p)
						}
						eggPath.Close()

						eggArea := clip.Outline{Path: eggPath.End()}.Op()
						color := color.NRGBA{R: 255, G: uint8(239 * (1 - progress)), B: uint8(174 * (1 - progress)), A: 255}
						paint.FillShape(gtx.Ops, color, eggArea)
						d := image.Point{Y: 375}

						return layout.Dimensions{Size: d}
					},
				),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					ed := material.Editor(th, &boilDurationInput, "sec")
					boilDurationInput.SingleLine = true
					boilDurationInput.Alignment = text.Middle

					if boiling && progress < 1 {
						boilRemain := (1 - progress) * float32(anim.duration.Seconds())
						// Format to 1 decimal.
						inputStr := fmt.Sprintf("%.1f", math.Round(float64(boilRemain)*10)/10)
						// Update the text in the inputbox
						boilDurationInput.SetText(inputStr)
					}

					margins := layout.Inset{
						Top:    unit.Dp(0),
						Right:  unit.Dp(170),
						Bottom: unit.Dp(40),
						Left:   unit.Dp(170),
					}

					border := widget.Border{
						Color:        color.NRGBA{R: 204, G: 204, B: 204, A: 255},
						CornerRadius: unit.Dp(3),
						Width:        unit.Dp(2),
					}

					return margins.Layout(gtx,
						func(gtx layout.Context) layout.Dimensions {
							return border.Layout(gtx, ed.Layout)
						},
					)
				}),

				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					bar := material.ProgressBar(th, progress)
					if boiling && progress < 1 {
						inv := op.InvalidateCmd{At: gtx.Now.Add(time.Second / 25)}
						gtx.Execute(inv)
					}
					return bar.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {

					// TWO: ... then we lay out those margins ...
					return layout.UniformInset(unit.Dp(25)).Layout(gtx,
						func(gtx layout.Context) layout.Dimensions {
							var text string
							if !boiling {
								text = "Start"
							} else {
								text = "Stop"
							}

							if boiling && progress >= 1 {
								text = "Finished"
							}

							btn := material.Button(th, &startButton, text)
							return btn.Layout(gtx)
						},
					)
				}),
			)

			evt.Frame(gtx.Ops)
		case app.DestroyEvent:
			return err
		}
	}

}

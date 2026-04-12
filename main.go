package main

import (
	"fmt"
	"image/color"
	"os"
	"unsafe"

	"github.com/sqweek/dialog"
	"golang.org/x/sys/windows"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

var (
	user32                   = windows.NewLazySystemDLL("user32.dll")
	getForegroundWindow      = user32.NewProc("GetForegroundWindow")
	getWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
)

func getCallerProcessName() (string, error) {
	// Get the foreground window handle
	hwnd, _, _ := getForegroundWindow.Call()
	if hwnd == 0 {
		return "no window", fmt.Errorf("no foreground window")
	}

	// Get the process ID from the window handle
	var pid uint32
	getWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))

	// Open the process and get its executable path
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return "open proc error", err
	}
	defer windows.CloseHandle(handle)

	buf := make([]uint16, 260)
	size := uint32(len(buf))
	err = windows.QueryFullProcessImageName(handle, 0, &buf[0], &size)
	if err != nil {
		return "windows.QueryFullProcessImageName", err
	}

	return windows.UTF16ToString(buf[:size]), nil
}

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

	args := os.Args // os.Args[0] = path to your exe, os.Args[1] = the URL

	url := ""
	caller, _ := getCallerProcessName()

	if len(args) > 1 {
		url = os.Args[1]
	}
	// url := os.Args[1]

	fmt.Println(args)
	// fmt.Println(url)

	go func() {
		// create new window
		w := new(app.Window)

		w.Option(app.Title("Link Intercept"))
		w.Option(app.Size(unit.Dp(400), unit.Dp(600)))
		// ops are the operations from the UI
		var ops op.Ops

		// startButton is a clickable widget
		var startButton widget.Clickable
		// var startButton2 widget.Clickable

		// startButton.Clicked()
		// th defines the material design style
		th := material.NewTheme()
		// listen for events in the window

		// clicked := false
		for {
			// first grab the event
			evt := w.Event()

			// then detect the type
			switch typ := evt.(type) {
			// this is sent when the application should re-render.
			case app.FrameEvent:
				// fmt.Println("----FrameEvent")
				gtx := app.NewContext(&ops, typ)

				if startButton.Clicked(gtx) {
					// clicked = true
					openDialog()
					fmt.Println("Button clicked1!")
				}

				layout.Flex{
					// Vertical alignment, from top to bottom
					Axis: layout.Vertical,
					// Empty space is left at the start, i.e. at the top
					Spacing: layout.SpaceStart,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(th, unit.Sp(20), caller)
						lbl.Color = color.NRGBA{R: 50, G: 50, B: 50, A: 255}
						lbl.LineHeight = unit.Sp(32)
						lbl.Alignment = text.Middle
						return lbl.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(th, unit.Sp(20), url)
						lbl.Color = color.NRGBA{R: 50, G: 50, B: 50, A: 255}
						lbl.LineHeight = unit.Sp(32)
						lbl.Alignment = text.Middle
						return lbl.Layout(gtx)
					}),
					// We insert two rigid elements:
					// First one to hold a button ...
					layout.Rigid(
						func(gtx layout.Context) layout.Dimensions {
							btn := material.Button(th, &startButton, "Start")
							// btn.Inset.Top = unit.Dp(40)
							return btn.Layout(gtx)
						},
					),
					// ... then one to hold an empty spacer
					layout.Rigid(
						// The height of the spacer is 25 Device independent pixels
						layout.Spacer{Height: unit.Dp(25)}.Layout,
					),
				)

				// typ.Source

				// btn := material.Button(th, &startButton, "Start")
				// fmt.Println("New Draw")
				// btn.Layout(gtx)

				typ.Frame(gtx.Ops)

			// and this is sent when the application should exits
			case app.DestroyEvent:
				os.Exit(0)
			}
		}
	}()

	app.Main()
}

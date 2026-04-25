package main

import (
	"log"
	"os"

	"github.com/crispyarty/LinkInterceptor/internal/system"
	"github.com/crispyarty/LinkInterceptor/internal/ui"
	"github.com/crispyarty/LinkInterceptor/internal/ui/dispatch"
	"github.com/crispyarty/LinkInterceptor/internal/ui/state"
	"github.com/crispyarty/LinkInterceptor/internal/ui/uicore"

	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/op"
)

func main() {
	go func() {
		w := new(app.Window)

		if err := run(w); err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}()

	app.Main()
}

func run(window *app.Window) (err error) {
	appUi := ui.NewApp(window)

	var ops op.Ops
	var eventCh = make(chan event.Event)
	var acks = make(chan struct{})

	go fillState(appUi.State())

	go func() {
		for {
			evt := window.Event()

			eventCh <- evt
			<-acks

			if _, ok := evt.(app.DestroyEvent); ok {
				return
			}
		}
	}()

	for {
		select {
		case evt := <-eventCh:
			switch evt := evt.(type) {
			case app.FrameEvent:
				gtx := uicore.Context{Context: app.NewContext(&ops, evt), App: appUi}
				// gtx := app.NewContext(&ops, evt)
				// gtx = uicore.WithApp(gtx, appUi)

				appUi.Layout(gtx)
				evt.Frame(gtx.Ops)
			case app.DestroyEvent:
				acks <- struct{}{}
				return evt.Err
			}
			acks <- struct{}{}
		case updateUi := <-dispatch.Actions:
			updateUi()
			window.Invalidate()
		}
	}
}

func fillState(state *state.State) {
	if browsers, err := system.GetBrowsers(); err == nil {
		// browsers = append(browsers, browsers...)
		// browsers = append(browsers, browsers...)
		// browsers = append(browsers, browsers...)
		state.Browsers.Items.Update(browsers)
	}
}

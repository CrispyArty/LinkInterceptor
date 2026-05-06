package ui

import (
	"os"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"

	// "github.com/crispyarty/LinkInterceptor/internal/system"

	"github.com/crispyarty/LinkInterceptor/internal/system"
	"github.com/crispyarty/LinkInterceptor/internal/ui/screens"
	"github.com/crispyarty/LinkInterceptor/internal/ui/state"
	"github.com/crispyarty/LinkInterceptor/internal/ui/uicore"
)

type App struct {
	theme  *material.Theme
	state  *state.State
	router *uicore.Router
}

func (a *App) State() *state.State {
	return a.state
}
func (a *App) Theme() *material.Theme {
	return a.theme
}
func (a *App) Router() *uicore.Router {
	return a.router
}

func NewApp(window *app.Window) *App {
	winWidth := 400
	winHeight := 600
	// screenWidth, screenHeight := system.GetScreenSize()

	// x := (screenWidth - winWidth) / 2
	// y := (screenHeight - winHeight) / 2

	window.Option(
		app.Title("Link Interceptor"),
		app.Size(unit.Dp(winWidth), unit.Dp(winHeight)),
	)

	args := os.Args // os.Args[0] = path to myapp exe, os.Args[1] = the URL

	s := state.NewState()

	if len(args) > 1 {
		s.Url = os.Args[1]
	}

	if caller, err := system.GetCaller(); err == nil {
		s.Caller = caller
	}

	return &App{
		state: s,
		theme: material.NewTheme(),
		router: uicore.NewRouter(
			uicore.Routes.Home,
			map[uicore.RouteID]uicore.ScreenBuilder{
				uicore.Routes.Home:     func() uicore.Screen { return screens.NewHome(s) },
				uicore.Routes.Settings: func() uicore.Screen { return nil },
			},
		),
	}
}

func (app *App) Layout(gtx uicore.Context) layout.Dimensions {
	// return layout.Flex{
	// 	Axis:    layout.Vertical,
	// 	Spacing: layout.SpaceStart,
	// }.Layout(gtx.Context)

	// gtx.App.Router().Layout(gtx)

	// fmt.Println(app.State().Url)

	return app.router.Layout(gtx)
}

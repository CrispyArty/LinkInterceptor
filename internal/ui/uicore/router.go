package uicore

import (
	"gioui.org/layout"
)

type RouteID string

var Routes = struct {
	Home     RouteID
	Browsers RouteID
	Settings RouteID
}{
	Home:     "home",
	Settings: "settings",
}

type Screen interface {
	Layout(gtx Context) layout.Dimensions
	Destroy()
}

type ScreenBuilder func() Screen

type Router struct {
	current Screen
	routes  map[RouteID]ScreenBuilder
}

func NewRouter(defaultRoute RouteID, registrty map[RouteID]ScreenBuilder) *Router {
	return &Router{
		current: registrty[defaultRoute](),
		routes:  registrty,
	}
}

func (app *Router) Layout(gtx Context) layout.Dimensions {
	if app.current == nil {
		return layout.Dimensions{}
	}
	return app.current.Layout(gtx)
}

func (r *Router) NavigateTo(next Screen) {
	if r.current != nil {
		r.current.Destroy()
	}
	r.current = next
}

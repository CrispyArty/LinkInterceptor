package uicore

import (
	"gioui.org/layout"
	"gioui.org/widget/material"
	"github.com/crispyarty/LinkInterceptor/internal/ui/state"
)

type AppProvider interface {
	State() *state.State
	Theme() *material.Theme
	Router() *Router
}

// type AppProvider interface {
// 	State() *state.State
// 	Theme() *material.Theme
// 	Router() *Router
// }

type Context struct {
	layout.Context
	// Theme    *material.Theme
	// AppState *state.State
	// Router   *Router
	App AppProvider
}

func (c Context) With(gtx layout.Context) Context {
	c.Context = gtx
	return c
}

// const appKey string = "theme"

// func WithApp(gtx layout.Context, appUi AppProvider) layout.Context {
// 	if gtx.Values == nil {
// 		gtx.Values = make(map[string]any)
// 	}
// 	gtx.Values[appKey] = appUi

// 	return gtx
// }

// func GetApp(gtx layout.Context) AppProvider {
// 	if app, ok := gtx.Values[appKey].(AppProvider); ok {
// 		return app
// 	}
// 	panic("App not found in context!")
// }

// const themeKey string = "theme"
// const stateKey string = "theme"
// const routerKey string = "theme"

// func WithTheme(gtx layout.Context, th *material.Theme) layout.Context {
// 	if gtx.Values == nil {
// 		gtx.Values = make(map[string]any)
// 	}
// 	gtx.Values[string(themeKey)] = th

// 	return gtx
// }

// func WithRouter(gtx layout.Context, s *state.State) layout.Context {
// 	if gtx.Values == nil {
// 		gtx.Values = make(map[string]any)
// 	}
// 	gtx.Values[string(routerKey)] = s

// 	return gtx
// }

// func WithState(gtx layout.Context, s *Router) layout.Context {
// 	if gtx.Values == nil {
// 		gtx.Values = make(map[string]any)
// 	}
// 	gtx.Values[string(stateKey)] = s

// 	return gtx
// }

// func GetTheme(gtx layout.Context) *material.Theme {
// 	if th, ok := gtx.Values[themeKey].(*material.Theme); ok {
// 		return th
// 	}
// 	panic("Theme not found in context!")
// }

// func GetState(gtx layout.Context) *state.State {
// 	if th, ok := gtx.Values[themeKey].(*state.State); ok {
// 		return th
// 	}
// 	panic("State not found in context!")
// }

// func GetRouter(gtx layout.Context) *Router {
// 	if th, ok := gtx.Values[themeKey].(*Router); ok {
// 		return th
// 	}
// 	panic("Router not found in context!")
// }

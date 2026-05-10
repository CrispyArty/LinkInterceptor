package screens

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/crispyarty/LinkInterceptor/internal/ui/components"
	"github.com/crispyarty/LinkInterceptor/internal/ui/state"
	"github.com/crispyarty/LinkInterceptor/internal/ui/uicore"
)

type Home struct {
	BrowsersList *components.BrowsersList
	DownloadCard *components.DownloadCard
	destroyers   []func()
}

func NewHome(s *state.State) *Home {
	dList, dDestroy := components.NewDownloadCard(s)
	bList, bDestroy := components.NewBrowsersList(s)

	return &Home{
		BrowsersList: bList,
		DownloadCard: dList,
		// DownloadCard: &components.DownloadCard{},
		destroyers: []func(){bDestroy, dDestroy},
	}
}

func (h *Home) Destroy() {
	for _, destroyer := range h.destroyers {
		destroyer()
	}
}

func (h *Home) Layout(mgtx uicore.Context) layout.Dimensions {
	return layout.UniformInset(unit.Dp(4)).Layout(mgtx.Context, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:    layout.Vertical,
			Spacing: layout.SpaceStart,
		}.Layout(gtx,
			// layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			// 	btn := material.Button(theme, &h.button, "Download")
			// 	// btn.Inset.Top = unit.Dp(40)
			// 	return btn.Layout(gtx)
			// }),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return h.DownloadCard.Layout(mgtx.With(gtx))
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),

			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {

				return h.BrowsersList.Layout(mgtx.With(gtx))
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(24)}.Layout),
		)
	})
}

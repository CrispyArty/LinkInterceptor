package components

import (
	"fmt"
	"log"
	"path/filepath"
	"sync/atomic"
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/crispyarty/LinkInterceptor/internal/fetcher"
	"github.com/crispyarty/LinkInterceptor/internal/system"
	"github.com/crispyarty/LinkInterceptor/internal/ui/dispatch"
	"github.com/crispyarty/LinkInterceptor/internal/ui/state"
	"github.com/crispyarty/LinkInterceptor/internal/ui/uicore"
)

type DownloadCard struct {
	card            Card
	list            widget.List
	url             string
	urlInfo         *fetcher.UrlInfo
	progressBytes   int64
	progressPercent float64
	downloading     bool
	downloaded      bool
	downloadError   error
	saveBtn         widget.Clickable
	saveAsBtn       widget.Clickable
}

func fetchInfo(d *DownloadCard) {
	urlInfo, _ := fetcher.GetHeaders(d.url)
	fmt.Println(urlInfo)
	dispatch.Actions <- func() {
		d.urlInfo = urlInfo
	}
}

func NewDownloadCard(s *state.State) (*DownloadCard, func()) {
	d := &DownloadCard{
		card: DefaultCard,
		url:  s.Url,
	}

	d.list.List.Axis = layout.Vertical

	go fetchInfo(d)

	// unsub := s.Browsers.Items.Subscribe(l.HandleData)

	return d, func() {}
}

func _saveUrlTo(url, dest string, onProgress func(bytes int64, percent float64), onDone func(err error)) {
	status, errs := fetcher.StartDownload(url, dest)

	go func() {
		ticker := time.NewTicker(time.Second / 60)
		defer ticker.Stop()

		var isUpdating atomic.Bool

		var updateProgress = func(force bool) {
			if !force && isUpdating.Load() {
				return
			}

			bytes := status.ProgressBytes.Load()
			percent := status.CalcPercent(bytes)

			isUpdating.Store(true)

			dispatch.Actions <- func() {
				defer isUpdating.Store(false)
				onProgress(bytes, percent)
			}
		}

		for {
			select {
			case <-ticker.C:
				updateProgress(false)
			case err := <-errs:
				if err != nil {
					log.Printf("Download failed: %v", err)
				}
				updateProgress(true)
				dispatch.Actions <- func() {
					onDone(err)
				}

				return
			}
		}
	}()

}

func saveUrlTo(url, dest string, ui *DownloadCard) {
	ui.downloading = true
	dispatch.Actions <- func() {} // empty action for window invalidate
	_saveUrlTo(url, dest,
		func(bytes int64, percent float64) {
			ui.progressBytes = bytes
			ui.progressPercent = percent
		},
		func(err error) {
			ui.downloadError = err
			ui.progressPercent = 100
			ui.downloaded = true
			ui.downloading = false
		},
	)
}

func (t *DownloadCard) handleSaveAsClick(url string, urlInfo *fetcher.UrlInfo) {
	if dir, err := system.OpenDialog(urlInfo.Filename); err == nil {
		saveUrlTo(t.url, dir, t)
	}
}

func (t *DownloadCard) handleSaveClick(url string, urlInfo *fetcher.UrlInfo) {
	if dir, err := system.GetDownloadsDir(); err == nil {
		saveUrlTo(t.url, filepath.Join(dir, urlInfo.Filename), t)
	}
}

func (t *DownloadCard) loading(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	return material.Body1(theme, "Loading...").Layout(gtx)
}

func (t *DownloadCard) doneActions(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	return layout.Flex{
		Axis:    layout.Vertical,
		Spacing: layout.SpaceStart,
	}.Layout(gtx,
	 	layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if !t.downloading {
				// return layout.Dimensions{Size: image.Point{Y: 20}}
				return layout.Dimensions{} // empty space before downloading begin
			}
			bar := material.ProgressBar(theme, float32(t.progressPercent/100))
			return bar.Layout(gtx)
		}),
	 )
}

func (t *DownloadCard) downloadActions(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	if t.saveBtn.Clicked(gtx) {
		t.handleSaveClick(t.url, t.urlInfo)
	}

	if t.saveAsBtn.Clicked(gtx) {
		t.handleSaveAsClick(t.url, t.urlInfo)
	}

	return layout.Flex{
		Axis:    layout.Vertical,
		Spacing: layout.SpaceStart,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if !t.downloading {
				// return layout.Dimensions{Size: image.Point{Y: 20}}
				return layout.Dimensions{} // empty space before downloading begin
			}
			bar := material.ProgressBar(theme, float32(t.progressPercent/100))
			return bar.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Alignment: layout.Middle,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					btn := material.Button(theme, &t.saveBtn, "Save")
					return btn.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),

				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					btn := material.Button(theme, &t.saveAsBtn, "Save As")
					return btn.Layout(gtx)
				}),
			)
		}),
	)
}

func (t *DownloadCard) regularLinkHint(gtx layuot.Context, theme *material.Theme) layout.Dimensions {
	return material.Body1(theme, "This is just a regular link").Laout(gtx)
}

func (t *DownloadCard) Layout(mgtx uicore.Context) layout.Dimensions {
	theme := mgtx.App.Theme()

	return t.card.Layout(mgtx.Context, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:    layout.Vertical,
			Spacing: layout.SpaceStart,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return material.Body1(theme, fmt.Sprintf("Link: %v", t.url)).Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				
				if t.urlInfo == nil {
					return t.loading(gtx, theme)
				} else if t.downloaded {
					return t.doneActions(gtx, theme)
				} else if t.urlInfo.Downloadable {
					return t.downloadActions(gtx, theme)
				} else {
					return t.regularLinkHint(gtx, theme)
				}
			}),
		)	
	})
}

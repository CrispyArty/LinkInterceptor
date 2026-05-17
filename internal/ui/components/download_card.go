package components

import (
	"io"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"gioui.org/io/clipboard"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/crispyarty/LinkInterceptor/internal/fetcher"
	"github.com/crispyarty/LinkInterceptor/internal/system"
	"github.com/crispyarty/LinkInterceptor/internal/ui/assets"
	"github.com/crispyarty/LinkInterceptor/internal/ui/dispatch"
	"github.com/crispyarty/LinkInterceptor/internal/ui/state"
	"github.com/crispyarty/LinkInterceptor/internal/ui/uicore"
)

type FileStatus struct {
	progressBytes   int64
	progressPercent float64
	downloading     bool
	downloaded      bool
	downloadError   error
	downloadedPath  string
}

type DownloadCard struct {
	card    Card
	list    widget.List
	url     string
	urlInfo *fetcher.UrlInfo

	file FileStatus

	saveBtn     widget.Clickable
	saveAsBtn   widget.Clickable
	openBtn     widget.Clickable
	copyPathBtn widget.Clickable
	copyUrlBtn  widget.Clickable
	linkEditor  widget.Editor
	copyIcon    *widget.Icon
}

func fetchInfo(d *DownloadCard) {
	urlInfo, _ := fetcher.GetHeaders(d.url)
	dispatch.Actions <- func() {
		d.urlInfo = urlInfo
	}
}

func NewDownloadCard(s *state.State) (*DownloadCard, func()) {
	d := &DownloadCard{
		card: DefaultCard,
		url:  s.Url,
	}

	d.copyIcon, _ = widget.NewIcon(assets.IconClipboard)

	d.linkEditor.ReadOnly = true
	// d.linkEditor.SingleLine = true
	d.linkEditor.SetText(d.url)

	d.list.List.Axis = layout.Vertical

	go fetchInfo(d)

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
	ui.file.downloading = true
	dispatch.Actions <- func() {} // empty action for window invalidate
	_saveUrlTo(url, dest,
		func(bytes int64, percent float64) {
			ui.file.progressBytes = bytes
			ui.file.progressPercent = percent
		},
		func(err error) {
			ui.file.downloadedPath = dest
			ui.file.downloadError = err
			ui.file.progressPercent = 100
			ui.file.downloaded = true
			ui.file.downloading = false
		},
	)
}

func (t *DownloadCard) loading(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	return material.Body1(theme, "Loading...").Layout(gtx)
}

func handleOpenFolderClick(dest string) {
	exec.Command("explorer", "/select,", dest).Run()
}

func copyToClipboard(gtx layout.Context, path string) {
	gtx.Execute(clipboard.WriteCmd{
		Type: "text/plain",
		Data: io.NopCloser(strings.NewReader(path)),
	})
}

func (t *DownloadCard) doneActions(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	if t.openBtn.Clicked(gtx) {
		handleOpenFolderClick(t.file.downloadedPath)
	}

	if t.copyPathBtn.Clicked(gtx) {
		copyToClipboard(gtx, t.file.downloadedPath)
	}

	return layout.Flex{}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			btn := material.Button(theme, &t.openBtn, "Open Folder")
			if t.openBtn.Hovered() {
				pointer.CursorPointer.Add(gtx.Ops)
			}

			return btn.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			btn := SecondaryButton(theme, &t.copyPathBtn, "Copy path")
			if t.copyPathBtn.Hovered() {
				pointer.CursorPointer.Add(gtx.Ops)
			}

			return btn.Layout(gtx)
		}),
	)
}

func (t *DownloadCard) handleSaveAsClick() {
	if dir, err := system.OpenDialog(t.urlInfo.Filename); err == nil {
		saveUrlTo(t.url, dir, t)
	}
}

func (t *DownloadCard) handleSaveClick() {
	if dir, err := system.GetDownloadsDir(); err == nil {
		saveUrlTo(t.url, filepath.Join(dir, t.urlInfo.Filename), t)
	}
}

func (t *DownloadCard) downloadActions(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	if t.saveBtn.Clicked(gtx) {
		t.handleSaveClick()
	}

	if t.saveAsBtn.Clicked(gtx) {
		t.handleSaveAsClick()
	}

	return layout.Flex{
		Axis:    layout.Vertical,
		Spacing: layout.SpaceStart,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if !t.file.downloading {
				return layout.Dimensions{}
			}
			bar := material.ProgressBar(theme, float32(t.file.progressPercent/100))

			return layout.Inset{Bottom: unit.Dp(8)}.Layout(gtx, bar.Layout)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if !t.file.downloading {
				return layout.Dimensions{}
			}

			return layout.Spacer{Height: unit.Dp(8)}.Layout(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					btn := material.Button(theme, &t.saveBtn, "Save")
					if t.saveBtn.Hovered() {
						pointer.CursorPointer.Add(gtx.Ops)
					}

					return btn.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					btn := SecondaryButton(theme, &t.saveAsBtn, "Save As")
					if t.saveAsBtn.Hovered() {
						pointer.CursorPointer.Add(gtx.Ops)
					}
					return btn.Layout(gtx)
				}),
			)
		}),
	)
}

func (t *DownloadCard) regularLinkHint(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	return material.Body1(theme, "This is just a regular link").Layout(gtx)
}

func (t *DownloadCard) Layout(mgtx uicore.Context) layout.Dimensions {
	theme := mgtx.App.Theme()

	return t.card.Layout(mgtx.Context, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:    layout.Vertical,
			Spacing: layout.SpaceStart,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis: layout.Horizontal,
				}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return material.Editor(theme, &t.linkEditor, "").Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if t.copyUrlBtn.Clicked(gtx) {
							copyToClipboard(gtx, t.url)
						}

						if t.copyUrlBtn.Hovered() {
							pointer.CursorPointer.Add(gtx.Ops)
						}

						btn := material.ButtonLayout(theme, &t.copyUrlBtn)
						btn.Background = uicore.Colors.ButtonIconBg

						return btn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return t.copyIcon.Layout(gtx, uicore.Colors.Black)
							})
						})
					}),
				)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if t.urlInfo == nil {
					return t.loading(gtx, theme)
				} else if t.file.downloaded {
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

// func drawOverflowFade(gtx layout.Context) layout.Dimensions {
// 	fadeWidth := gtx.Dp(24)
// 	size := gtx.Constraints.Min

// 	// Create a rectangle for the fade area at the right edge
// 	rect := image.Rect(size.X-fadeWidth, 0, size.X, size.Y)

// 	// Define the gradient
// 	// Transparent -> Background Color
// 	cols := [2]color.NRGBA{
// 		{R: 255, G: 255, B: 255, A: 0},   // Transparent
// 		{R: 255, G: 255, B: 255, A: 255}, // White (match your BG)
// 	}

// 	// Setup the gradient operation
// 	grad := paint.LinearGradientOp{
// 		Stop1:  layout.FPt(image.Pt(rect.Min.X, 0)),
// 		Color1: cols[0],
// 		Stop2:  layout.FPt(image.Pt(rect.Max.X, 0)),
// 		Color2: cols[1],
// 	}

// 	// Apply the gradient to the specific area
// 	defer clip.Rect(rect).Push(gtx.Ops).Pop()
// 	grad.Add(gtx.Ops)
// 	paint.PaintOp{}.Add(gtx.Ops)

// 	return layout.Dimensions{Size: size}
// }

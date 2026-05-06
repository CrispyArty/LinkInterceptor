package components

import (
	"fmt"
	"image"
	"os/exec"
	"strings"
	"sync/atomic"

	"gioui.org/layout"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/crispyarty/LinkInterceptor/internal/system"
	"github.com/crispyarty/LinkInterceptor/internal/ui/dispatch"
	"github.com/crispyarty/LinkInterceptor/internal/ui/state"
	"github.com/crispyarty/LinkInterceptor/internal/ui/uicore"
)

type BrowserItem struct {
	path, name string
	shellPath  string
	icon       paint.ImageOp
	data       system.Browser
	button     widget.Clickable
	card       CardButton
	// onClick    func()
}

type BrowsersList struct {
	caller   string
	url      string
	version  atomic.Int64
	browsers []*BrowserItem
	cache    map[string]*BrowserItem
	list     widget.List
}

func (t *BrowserItem) Update(data *system.Browser) {
	t.path = data.Path
	t.name = data.Name
	t.shellPath = data.ShellPath
	// t.icon = &(*data.IconImageOp)
	t.icon = paint.NewImageOp(data.IconImage)
}

// old
// [{path: 'firefox.exe'}, {path: 'chrome.exe'}, {path: 'ie.exe'}]
// new
// [{path: 'firefox.exe'}, {path: 'chrome.exe'}, {path: 'ie.exe'}, {path: 'arc.exe'}]
func (t *BrowsersList) HandleData(data []*system.Browser) {
	// process data in goroutine
	v := t.version.Add(1)

	go func(currentVersion int64) {
		// transform data to uiElemetns
		browsers, cache := uicore.SyncList(
			data,
			t.cache,
			func(item *system.Browser) string { return item.Name },
			func() *BrowserItem { return &BrowserItem{} },
		)

		// dispatch to channel, receiving end will call w.Invalidate() to draw with new data
		dispatch.Actions <- func() {
			if currentVersion == t.version.Load() {
				t.browsers, t.cache = browsers, cache
			}
		}
	}(v)
}

func NewBrowsersList(s *state.State) (*BrowsersList, func()) {
	fmt.Println(s.Url)

	l := &BrowsersList{
		url:    s.Url,
		caller: s.Caller,
		cache:  make(map[string]*BrowserItem),
	}

	l.list.List.Axis = layout.Vertical
	unsub := s.Browsers.Items.Subscribe(l.HandleData)

	return l, func() {
		unsub()
	}
}

func (t *BrowsersList) handleBrowserClick(bi *BrowserItem) {
	fmt.Printf("path: %v | %v | %v\n", bi.shellPath, t.url, t.caller)
	cmdStr := bi.shellPath
	if !strings.Contains(bi.shellPath, "%1") {
		cmdStr += " %1"
	}

	cmdStr = strings.Replace(cmdStr, "%1", "https://example.com", 1)

	parts := system.SplitWindowsArgs(cmdStr)

	// fmt.Printf("cmdStr: %v\n", cmdStr)
	// fmt.Printf("parts: %v\n", parts)
	// fmt.Printf("parts2: %v\n", system.SplitWindowsArgs(cmdStr))

	cmd := exec.Command(parts[0], parts[1:]...)

	// Start the process
	err := cmd.Start()

	// os.Exit(0)

	fmt.Println(err)

	fmt.Printf("cmd: %v\n", cmd)
}

func (t *BrowserItem) layout(gtx layout.Context, theme *material.Theme) layout.Dimensions {

	return t.card.Layout(gtx, &t.button, func(gtx layout.Context) layout.Dimensions {

		return layout.Flex{
			Alignment: layout.Middle,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				size := gtx.Dp(32)

				if t.icon.Size() == (image.Point{}) {
					return layout.Dimensions{Size: image.Point{X: size, Y: size}}
				}

				gtx.Constraints = layout.Exact(image.Point{X: size, Y: size})

				return widget.Image{
					Src: t.icon,
					Fit: widget.Contain,
				}.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return material.Body1(theme, t.name).Layout(gtx)
			}),
		)
	})
}

func (t *BrowsersList) Layout(mgtx uicore.Context) layout.Dimensions {
	theme := mgtx.App.Theme()

	// for _, b := range t.browsers {
	// 	if b.button.Clicked(mgtx.Context) {
	// 		t.handleBrowserClick(b)
	// 	}
	// }
	// if t.list.ScrollDistance() != 0 {
	// 	cmd := op.InvalidateCmd{}
	// 	mgtx.Execute(cmd)
	// }

	return material.List(theme, &t.list).Layout(mgtx.Context, len(t.browsers), func(gtx layout.Context, i int) layout.Dimensions {

		return layout.Inset{Bottom: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			if t.browsers[i].button.Clicked(gtx) {
				t.handleBrowserClick(t.browsers[i])
			}

			return t.browsers[i].layout(gtx, theme)
		})
	})
}

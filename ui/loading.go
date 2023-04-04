package ui

import (
	"github.com/linweiyuan/go-chatgpt/common"
	"github.com/rivo/tview"
)

func (ui *UI) StartLoading(view interface{}) {
	ui.app.QueueUpdateDraw(func() {
		view.(*tview.Box).SetTitle(common.LoadingText)
	})
}

func (ui *UI) StopLoading(view interface{}) {
	ui.app.QueueUpdateDraw(func() {
		view.(*tview.Box).SetTitle("")
	})
}

package ui

import "github.com/rivo/tview"

func (ui *UI) StartLoading(view interface{}) {
	ui.App.QueueUpdateDraw(func() {
		view.(*tview.Box).SetTitle("Loading...")
	})
}

func (ui *UI) StopLoading(view interface{}) {
	ui.App.QueueUpdateDraw(func() {
		view.(*tview.Box).SetTitle("")
	})
}

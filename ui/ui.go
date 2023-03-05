package ui

import (
	"github.com/rivo/tview"
)

type UI struct {
	App                      *tview.Application
	ConversationTreeNodeRoot *tview.TreeNode
	ConversationTreeView     *tview.TreeView
	ContentField             *tview.InputField
	MessageArea              *tview.TextArea
}

func New() *UI {
	return &UI{
		App: tview.NewApplication(),

		ConversationTreeNodeRoot: tview.NewTreeNode("Conversations"),
		ConversationTreeView:     tview.NewTreeView(),
		ContentField:             tview.NewInputField(),
		MessageArea:              tview.NewTextArea(),
	}
}

func (ui *UI) Setup() {
	ui.ConversationTreeView.SetRoot(ui.ConversationTreeNodeRoot).SetCurrentNode(ui.ConversationTreeNodeRoot)
	ui.ConversationTreeView.SetBorder(true)

	ui.ContentField.SetBorder(true)

	ui.MessageArea.SetBorder(true)

	rightFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	rightFlex.AddItem(ui.ContentField, 0, 1, false)
	rightFlex.AddItem(ui.MessageArea, 0, 9, false)

	mainFlex := tview.NewFlex()
	mainFlex.AddItem(ui.ConversationTreeView, 0, 4, false)
	mainFlex.AddItem(rightFlex, 0, 6, false)

	if err := ui.App.SetRoot(mainFlex, true).SetFocus(ui.ContentField).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

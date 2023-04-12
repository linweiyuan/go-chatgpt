package ui

import (
	"fmt"
	"github.com/linweiyuan/go-chatgpt/api"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/linweiyuan/go-chatgpt/common"
	"github.com/rivo/tview"
)

type UI struct {
	app *tview.Application

	conversationTreeNodeRoot *tview.TreeNode
	ConversationTreeView     *tview.TreeView
	contentField             *tview.InputField
	messageArea              *tview.TextArea

	api *api.API
}

func (ui *UI) Setup() {
	mainFlex := tview.NewFlex()

	modal := func(p tview.Primitive, width, height int) tview.Primitive {
		return tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(p, height, 1, true).
				AddItem(nil, 0, 1, false), width, 1, true).
			AddItem(nil, 0, 1, false)
	}

	ui.ConversationTreeView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyF2 {
			node := ui.ConversationTreeView.GetCurrentNode()
			level := node.GetLevel()
			if level == 0 || level == 2 {
				return nil
			}

			form := tview.NewForm().
				AddTextView("Current title", node.GetText(), 0, 2, true, false).
				AddInputField("New title", "", 0, nil, func(newTitle string) {
					node.SetText(newTitle)
				}).
				AddButton("Save", func() {
					go ui.renameTitle(node.GetText())
					ui.app.SetRoot(mainFlex, true).SetFocus(ui.ConversationTreeView)
				}).
				AddButton("Quit", func() {
					ui.app.SetRoot(mainFlex, true).SetFocus(ui.ConversationTreeView)
				})

			ui.app.SetRoot(modal(form, 40, 10), true)
		}

		if event.Key() == tcell.KeyCtrlR {
			node := ui.ConversationTreeView.GetCurrentNode()
			level := node.GetLevel()
			switch level {
			case 0:
				go ui.GetConversations()
			case 1:
				ui.renderConversationContent(node)
			}
		}

		return event
	})

	if !common.IsChatGPT {
		ui.contentField.SetTitleAlign(tview.AlignRight)
	}
	ui.contentField.SetBorder(true)
	ui.contentField.SetDoneFunc(func(key tcell.Key) {
		text := strings.TrimSpace(ui.contentField.GetText())
		if text == "" {
			return
		}

		if common.IsChatGPT {
			go ui.startConversation(text)
		} else {
			ui.messageArea.SetText("", false)
			go ui.chatCompletions(text)
		}
	})
	go func() {
		for {
			<-common.ConversationDoneChannel
			ui.contentField.SetText("")
		}
	}()

	ui.messageArea.SetBorder(true)
	ui.messageArea.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return nil
	})

	go func() {
		for responseText := range common.ResponseTextChannel {
			ui.app.QueueUpdateDraw(func() {
				if common.IsChatGPT {
					ui.messageArea.SetText(responseText, true)
				} else {
					ui.messageArea.SetText(ui.messageArea.GetText()+responseText, true)
				}
			})
		}
	}()

	rightFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	rightFlex.AddItem(ui.contentField, 3, 1, false)
	rightFlex.AddItem(ui.messageArea, 0, 9, false)

	if !common.IsChatGPT {
		ui.app.SetRoot(rightFlex, true).SetFocus(ui.contentField)
		return
	}

	ui.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			switch ui.app.GetFocus() {
			case ui.ConversationTreeView:
				ui.app.SetFocus(ui.contentField)
			case ui.contentField:
				ui.app.SetFocus(ui.ConversationTreeView)
			}
			return nil
		}
		return event
	})

	ui.ConversationTreeView.SetRoot(ui.conversationTreeNodeRoot).SetCurrentNode(ui.conversationTreeNodeRoot)
	ui.ConversationTreeView.SetBorder(true)
	ui.ConversationTreeView.SetSelectedFunc(func(node *tview.TreeNode) {
		reference := node.GetReference()
		if reference == nil {
			common.ConversationID = ""
			common.CurrentNode = nil
			ui.messageArea.SetText("", false)
			return
		}

		conversationItem, ok := node.GetReference().(common.ConversationItem)
		if ok {
			common.ConversationID = conversationItem.ID
		}

		if len(node.GetChildren()) == 0 {
			if !ui.renderConversationContent(node) {
				return
			}
		}
		ui.messageArea.SetText("", false)

		common.CurrentNode = node
	})

	mainFlex.AddItem(ui.ConversationTreeView, 0, 3, false)
	mainFlex.AddItem(rightFlex, 0, 7, false)

	go func() {
		for {
			<-common.ReloadConversationsChannel
			ui.GetConversations()
		}
	}()

	ui.app.SetRoot(mainFlex, true).SetFocus(ui.contentField)
}

func (ui *UI) renderConversationContent(node *tview.TreeNode) bool {
	conversationItem, ok := node.GetReference().(common.ConversationItem)
	if !ok {
		return false
	}

	go ui.getConversation(conversationItem.ID)

	go func(currentNode *tview.TreeNode) {
		currentNode.ClearChildren()

		for {
			select {
			case message := <-common.MessageChannel:
				questionTreeNode := tview.NewTreeNode(message.Content.Parts[0]).SetReference(message)
				questionTreeNode.SetSelectedFunc(func() {
					message := questionTreeNode.GetReference().(common.Message)
					ui.messageArea.SetText(common.QuestionAnswerMap[message.ID], true)
				})
				ui.app.QueueUpdateDraw(func() {
					currentNode.AddChild(questionTreeNode)
				})
			case <-common.ExitForLoopChannel:
				return
			}
		}
	}(node)

	return true
}

func New(api *api.API, app *tview.Application) *UI {
	if common.IsChatGPT {
		return &UI{
			app: app,

			conversationTreeNodeRoot: tview.NewTreeNode(common.NewChatText),
			ConversationTreeView:     tview.NewTreeView(),
			contentField:             tview.NewInputField(),
			messageArea:              tview.NewTextArea(),

			api: api,
		}
	}

	return &UI{
		app: app,

		contentField: tview.NewInputField(),
		messageArea:  tview.NewTextArea(),

		api: api,
	}
}

func (ui *UI) GetConversations() {
	ui.StartLoading(ui.ConversationTreeView.Box)
	defer ui.StopLoading(ui.ConversationTreeView.Box)

	conversations := ui.api.GetConversations()
	if conversations != nil {
		ui.renderConversationTree(conversations)
	}
}

func (ui *UI) renderConversationTree(conversations *common.Conversations) {
	ui.app.QueueUpdateDraw(func() {
		ui.conversationTreeNodeRoot.ClearChildren()

		var conversationID string
		if common.CurrentNode != nil {
			conversationItem, ok := common.CurrentNode.GetReference().(common.ConversationItem)
			if ok {
				conversationID = conversationItem.ID
			}
		}

		for _, conversation := range conversations.Items {
			conversationTreeNode := tview.NewTreeNode(conversation.Title).SetReference(conversation)
			if conversation.ID == conversationID {
				common.CurrentNode = conversationTreeNode
			}
			ui.conversationTreeNodeRoot.AddChild(conversationTreeNode)
		}

		if common.CurrentNode != nil {
			ui.ConversationTreeView.SetCurrentNode(common.CurrentNode)
		}
	})
}

func (ui *UI) getConversation(conversationItemID string) {
	ui.StartLoading(ui.ConversationTreeView.Box)
	defer ui.StopLoading(ui.ConversationTreeView.Box)

	ui.api.GetConversation(conversationItemID)
}

func (ui *UI) startConversation(text string) {
	ui.StartLoading(ui.messageArea.Box)
	defer ui.StopLoading(ui.messageArea.Box)

	ui.api.StartConversation(text)
}

func (ui *UI) renameTitle(text string) {
	ui.StartLoading(ui.ConversationTreeView.Box)
	defer ui.StopLoading(ui.ConversationTreeView.Box)

	ui.api.RenameTitle(text)
}

func (ui *UI) chatCompletions(text string) {
	ui.StartLoading(ui.messageArea.Box)
	defer ui.StopLoading(ui.messageArea.Box)

	ui.api.ChatCompletions(text)
}

func (ui *UI) CheckUsage() {
	checkUsageResponse := ui.api.CheckUsage()
	if checkUsageResponse != nil {
		ui.app.QueueUpdateDraw(func() {
			ui.contentField.SetTitle(fmt.Sprintf("Total Granted: %.2f | Total Used: %.2f | Total Available: %.2f",
				checkUsageResponse.TotalGranted,
				checkUsageResponse.TotalUsed,
				checkUsageResponse.TotalAvailable,
			))
		})
	}
}

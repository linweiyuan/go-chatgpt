package ui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/linweiyuan/go-chatgpt/api"
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

func New(api *api.API) *UI {
	return &UI{
		app: tview.NewApplication(),

		conversationTreeNodeRoot: tview.NewTreeNode("+ New chat"),
		ConversationTreeView:     tview.NewTreeView(),
		contentField:             tview.NewInputField(),
		messageArea:              tview.NewTextArea(),

		api: api,
	}
}

func (ui *UI) Setup() {
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
			conversationItem, ok := node.GetReference().(common.ConversationItem)
			if !ok {
				return
			}

			go ui.getConversation(conversationItem.ID)

			go func(currentNode *tview.TreeNode) {
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
		}
		ui.messageArea.SetText("", false)

		common.CurrentNode = node
	})

	ui.contentField.SetBorder(true)
	ui.contentField.SetDoneFunc(func(key tcell.Key) {
		text := strings.TrimSpace(ui.contentField.GetText())
		if text == "" {
			return
		}

		go ui.startConversation(text)
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
				ui.messageArea.SetText(responseText, true)
			})
		}
	}()

	rightFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	rightFlex.AddItem(ui.contentField, 3, 1, false)
	rightFlex.AddItem(ui.messageArea, 0, 9, false)

	mainFlex := tview.NewFlex()
	mainFlex.AddItem(ui.ConversationTreeView, 0, 3, false)
	mainFlex.AddItem(rightFlex, 0, 7, false)

	go func() {
		for {
			<-common.ReloadConversationsChannel
			ui.GetConversations()
		}
	}()

	if err := ui.app.SetRoot(mainFlex, true).SetFocus(ui.contentField).EnableMouse(true).Run(); err != nil {
		panic(err)
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

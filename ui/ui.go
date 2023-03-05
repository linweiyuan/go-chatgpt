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

		conversationTreeNodeRoot: tview.NewTreeNode("Conversations"),
		ConversationTreeView:     tview.NewTreeView(),
		contentField:             tview.NewInputField(),
		messageArea:              tview.NewTextArea(),

		api: api,
	}
}

func (ui *UI) Setup() {
	ui.ConversationTreeView.SetRoot(ui.conversationTreeNodeRoot).SetCurrentNode(ui.conversationTreeNodeRoot)
	ui.ConversationTreeView.SetBorder(true)
	ui.ConversationTreeView.SetSelectedFunc(func(node *tview.TreeNode) {
		conversationItem := node.GetReference()
		if conversationItem == nil {
			return
		}

		if len(node.GetChildren()) == 0 {
			conversationItem, ok := node.GetReference().(common.ConversationItem)
			if !ok {
				return
			}

			go ui.getConversation(conversationItem.ID, node)

			go func() {
				for {
					select {
					case nodeMessageMap := <-common.NodeMessageChannel:
						currentNode := nodeMessageMap[common.KEY_NODE].(*tview.TreeNode)
						message := nodeMessageMap[common.KEY_MESSAGE].(common.Message)

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
			}()
		} else {
			node.SetExpanded(!node.IsExpanded())
		}
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

			go ui.GetConversations()
		}
	}()

	ui.messageArea.SetBorder(true)
	go func() {
		for responseText := range common.ResponseTextChannel {
			ui.app.QueueUpdateDraw(func() {
				ui.messageArea.SetText(responseText, true)
			})
		}
	}()

	rightFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	rightFlex.AddItem(ui.contentField, 0, 1, false)
	rightFlex.AddItem(ui.messageArea, 0, 9, false)

	mainFlex := tview.NewFlex()
	mainFlex.AddItem(ui.ConversationTreeView, 0, 4, false)
	mainFlex.AddItem(rightFlex, 0, 6, false)

	if err := ui.app.SetRoot(mainFlex, true).SetFocus(ui.contentField).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

func (ui *UI) renderConversationTree(conversations *common.Conversations) {
	ui.app.QueueUpdateDraw(func() {
		ui.conversationTreeNodeRoot.ClearChildren()

		for _, conversation := range conversations.Items {
			conversationTreeNode := tview.NewTreeNode(conversation.Title).SetReference(conversation)
			ui.conversationTreeNodeRoot.AddChild(conversationTreeNode)
		}
	})
}

func (ui *UI) GetConversations() {
	ui.StartLoading(ui.ConversationTreeView.Box)
	defer ui.StopLoading(ui.ConversationTreeView.Box)

	conversations := ui.api.GetConversations()
	ui.renderConversationTree(conversations)
}

func (ui *UI) getConversation(conversationItemID string, node *tview.TreeNode) {
	ui.StartLoading(ui.ConversationTreeView.Box)
	defer ui.StopLoading(ui.ConversationTreeView.Box)

	ui.api.GetConversation(conversationItemID, node)
}

func (ui *UI) startConversation(text string) {
	ui.StartLoading(ui.messageArea.Box)
	defer ui.StopLoading(ui.messageArea.Box)

	ui.api.StartConversation(text)
}

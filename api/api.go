package api

import (
	"encoding/json"
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/linweiyuan/go-chatgpt/ui"
	"github.com/rivo/tview"
)

const API_SERVER_URL = "https://api.linweiyuan.com/chatgpt"

var (
	client *resty.Client
)

type Conversations struct {
	Items  []ConversationItem `json:"items"`
	Total  int                `json:"total"`
	Limit  int                `json:"limit"`
	Offset int                `json:"offset"`
}

type ConversationItem struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	CreateTime string `json:"create_time"`
}

func init() {
	client = resty.New().SetBaseURL(API_SERVER_URL)
	client.SetHeader("Authorization", os.Getenv("ACCESS_TOKEN"))
}

func GetConversations(ui *ui.UI) {
	ui.StartLoading(ui.ConversationTreeView.Box)
	defer ui.StopLoading(ui.ConversationTreeView.Box)

	resp, err := client.R().Get("/conversations")
	if err != nil {
		return
	}

	var conversations Conversations
	json.Unmarshal(resp.Body(), &conversations)

	ui.App.QueueUpdateDraw(func() {
		for _, conversation := range conversations.Items {
			conversationTreeNode := tview.NewTreeNode(conversation.Title).SetReference(conversation)
			ui.ConversationTreeNodeRoot.AddChild(conversationTreeNode)
		}
	})
}

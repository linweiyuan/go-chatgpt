package main

import (
	"github.com/linweiyuan/go-chatgpt/api"
	"github.com/linweiyuan/go-chatgpt/ui"
)

func main() {
	api := api.New()
	ui := ui.New(api)

	go func() {
		ui.StartLoading(ui.ConversationTreeView.Box)
		defer ui.StopLoading(ui.ConversationTreeView.Box)

		conversations := api.GetConversations()
		ui.RenderConversationTree(conversations)
	}()

	ui.Setup()
}

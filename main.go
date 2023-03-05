package main

import (
	"github.com/linweiyuan/go-chatgpt/api"
	"github.com/linweiyuan/go-chatgpt/ui"
)

func main() {
	api := api.New()
	ui := ui.New(api)
	go ui.GetConversations()
	ui.Setup()
}

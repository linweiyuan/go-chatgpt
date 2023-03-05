package main

import (
	"github.com/linweiyuan/go-chatgpt/api"
	"github.com/linweiyuan/go-chatgpt/ui"
)

func main() {
	ui := ui.New()
	go api.GetConversations(ui)
	ui.Setup()
}

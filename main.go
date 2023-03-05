package main

import (
	"github.com/linweiyuan/go-chatgpt/api"
	"github.com/linweiyuan/go-chatgpt/ui"
)

func main() {
	app := ui.New(api.New())
	go app.GetConversations()
	app.Setup()
}

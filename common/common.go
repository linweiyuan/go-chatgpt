package common

import "github.com/rivo/tview"

const (
	ChooseModeTitle      = "Which mode do you want to use?"
	ApiMode              = "API"
	ChatGPTMode          = "ChatGPT"
	LoadingText          = "Loading..."
	NewChatText          = "+ New chat"
	ChatGPTModel         = "text-davinci-002-render-sha"
	ApiModel             = "gpt-3.5-turbo"
	ApiVersion           = "/v1"
	RoleUser             = "user"
	RoleAssistant        = "assistant"
	CurrentTitle         = "Current title"
	NewTitle             = "New title"
	Save                 = "Save"
	Quit                 = "Quit"
	ClearAllConversation = "Do you want to clear all conversation?"
	Yes                  = "Yes"
	No                   = "No"
	DeleteConversation   = "Are you sure to to delete this conversation?"
)

var (
	QuestionAnswerMap = make(map[string]string)

	MessageChannel     = make(chan Message)
	ExitForLoopChannel = make(chan bool)

	ResponseTextChannel     = make(chan string)
	ConversationDoneChannel = make(chan bool)

	ParentMessageID string
	ConversationID  string

	ReloadConversationsChannel = make(chan bool)

	CurrentNode *tview.TreeNode

	IsChatGPT bool

	ApiMessages []ChatCompletionsMessage
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

type Conversation struct {
	CurrentNode string                        `json:"current_node"`
	Mapping     map[string]ConversationDetail `json:"mapping"`
	Title       string                        `json:"title"`
}

type ConversationDetail struct {
	ID       string   `json:"id"`
	Message  Message  `json:"message"`
	Parent   string   `json:"parent"`
	Children []string `json:"children"`
}

type Message struct {
	Author  Author  `json:"author"`
	Content Content `json:"content"`
	ID      string  `json:"id"`
	EndTurn bool    `json:"end_turn"`
}

type Author struct {
	Role string `json:"role"`
}

type Content struct {
	ContentType string   `json:"content_type"`
	Parts       []string `json:"parts"`
}

type CreateConversationResponse struct {
	ConversationID string  `json:"conversation_id"`
	Message        Message `json:"message"`
}

type ChatCompletionsRequest struct {
	Model    string                   `json:"model"`
	Messages []ChatCompletionsMessage `json:"messages"`
	Stream   bool                     `json:"stream"`
}

type ChatCompletionsMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionsResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Delta struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"delta"`
		Index        int    `json:"index"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

type CheckUsageResponse struct {
	TotalGranted   float32 `json:"total_granted"`
	TotalUsed      float64 `json:"total_used"`
	TotalAvailable float64 `json:"total_available"`
}

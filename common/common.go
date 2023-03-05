package common

const (
	KEY_NODE    = "node"
	KEY_MESSAGE = "message"
)

var (
	QuestionAnswerMap = make(map[string]string)

	NodeMessageChannel = make(chan map[string]interface{})
	ExitForLoopChannel = make(chan bool)

	ResponseTextChannel     = make(chan string)
	ConversationDoneChannel = make(chan bool)

	MessageID       string
	ParentMessageID string
	ConversationID  string
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
	Role    string  `json:"role"`
}

type Author struct {
	Role string `json:"role"`
}

type Content struct {
	ContentType string   `json:"content_type"`
	Parts       []string `json:"parts"`
}

type MakeConversationRequest struct {
	MessageID       string `json:"message_id"`
	ParentMessageID string `json:"parent_message_id"`
	ConversationID  string `json:"conversation_id"`
	Content         string `json:"content"`
}

type MakeConversationResponse struct {
	ConversationID string  `json:"conversation_id"`
	Message        Message `json:"message"`
}

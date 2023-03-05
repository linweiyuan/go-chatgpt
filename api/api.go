package api

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/linweiyuan/go-chatgpt/common"
	"github.com/rivo/tview"
)

const API_SERVER_URL = "https://api.linweiyuan.com/chatgpt"

var (
	client *resty.Client
)

type API struct {
}

func New() *API {
	return &API{}
}

func init() {
	client = resty.New().SetBaseURL(API_SERVER_URL)
	client.SetHeader("Authorization", os.Getenv("ACCESS_TOKEN"))
}

func (api *API) GetConversations() *common.Conversations {
	resp, err := client.R().Get("/conversations")
	if err != nil {
		return nil
	}

	var conversations common.Conversations
	json.Unmarshal(resp.Body(), &conversations)
	return &conversations
}

func (api *API) GetConversation(conversationID string, node *tview.TreeNode) {
	resp, err := client.R().Get("/conversation/" + conversationID)
	if err != nil {
		return
	}

	var conversation common.Conversation
	json.Unmarshal(resp.Body(), &conversation)

	handleConversationDetail(conversation.CurrentNode, conversation.Mapping, node)

	common.ExitForLoopChannel <- true
}

func handleConversationDetail(currentNode string, mapping map[string]common.ConversationDetail, node *tview.TreeNode) {
	conversationDetail := mapping[currentNode]
	parentID := conversationDetail.Parent
	if parentID != "" {
		common.QuestionAnswerMap[parentID] = strings.TrimSpace(conversationDetail.Message.Content.Parts[0])
		handleConversationDetail(parentID, mapping, node)
	}
	message := conversationDetail.Message
	parts := message.Content.Parts

	if len(parts) != 0 && parts[0] != "" {
		if message.Author.Role == "user" {
			common.NodeMessageChannel <- map[string]interface{}{
				common.KEY_NODE:    node,
				common.KEY_MESSAGE: message,
			}
		}
	}
}

package api

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
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

func (api *API) StartConversation(content string) {
	resp, err := client.R().
		SetDoNotParseResponse(true).
		SetHeader("Content-Type", "application/json").
		SetBody(common.MakeConversationRequest{
			MessageId:       uuid.NewString(),
			ParentMessageId: uuid.NewString(),
			ConversationId:  nil,
			Content:         content,
		}).
		Post("/conversation")
	if err != nil {
		return
	}

	defer resp.RawBody().Close()

	reader := bufio.NewReader(resp.RawBody())
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			common.ConversationDoneChannel <- true
			break
		}

		conversationDetail := parseEvent(line)
		if conversationDetail != nil {
			parts := conversationDetail.Message.Content.Parts
			if len(parts) != 0 {
				common.ResponseTextChannel <- parts[0]
			}
		}
	}
}

func parseEvent(line string) *common.ConversationDetail {
	if strings.Contains(line, "DONE") {
		return nil
	}

	if strings.HasPrefix(line, "data: ") {
		var conversationDetail common.ConversationDetail
		a := strings.TrimRight(strings.TrimPrefix(line, "data: "), "\n")
		json.Unmarshal([]byte(a), &conversationDetail)
		return &conversationDetail
	}

	return nil
}

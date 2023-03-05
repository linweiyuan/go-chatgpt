package api

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/linweiyuan/go-chatgpt/common"
)

var ServerUrl = os.Getenv("SERVER_URL")

var (
	client *resty.Client
)

type API struct {
}

func New() *API {
	return &API{}
}

func init() {
	client = resty.New().SetBaseURL(ServerUrl)
	client.SetHeader("Authorization", os.Getenv("ACCESS_TOKEN"))
}

func (api *API) GetConversations() *common.Conversations {
	resp, err := client.R().Get("/conversations")
	if err != nil {
		return nil
	}

	var conversations common.Conversations
	err = json.Unmarshal(resp.Body(), &conversations)
	if err != nil {
		return nil
	}

	return &conversations
}

func (api *API) GetConversation(conversationID string) {
	resp, err := client.R().Get("/conversation/" + conversationID)
	if err != nil {
		return
	}

	var conversation common.Conversation
	err = json.Unmarshal(resp.Body(), &conversation)
	if err != nil {
		return
	}

	currentNode := conversation.CurrentNode
	common.ParentMessageID = currentNode
	handleConversationDetail(currentNode, conversation.Mapping)

	common.ExitForLoopChannel <- true
}

func handleConversationDetail(currentNode string, mapping map[string]common.ConversationDetail) {
	conversationDetail := mapping[currentNode]
	parentID := conversationDetail.Parent
	if parentID != "" {
		common.QuestionAnswerMap[parentID] = strings.TrimSpace(conversationDetail.Message.Content.Parts[0])
		handleConversationDetail(parentID, mapping)
	}
	message := conversationDetail.Message
	parts := message.Content.Parts

	if len(parts) != 0 && parts[0] != "" {
		if message.Author.Role == "user" {
			common.MessageChannel <- message
		}
	}
}

var tempConversationID string

func (api *API) StartConversation(content string) {
	common.MessageID = uuid.NewString()
	parentMessageID := common.ParentMessageID
	if parentMessageID == "" || common.ConversationID == "" {
		parentMessageID = uuid.NewString()
	}
	resp, err := client.R().
		SetDoNotParseResponse(true).
		SetHeader("Content-Type", "application/json").
		SetBody(common.MakeConversationRequest{
			MessageID:       common.MessageID,
			ParentMessageID: parentMessageID,
			ConversationID:  common.ConversationID,
			Content:         content,
		}).
		Post("/conversation")
	if err != nil {
		return
	}

	// get it again from response
	common.ParentMessageID = ""

	defer func(body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			return
		}
	}(resp.RawBody())

	reader := bufio.NewReader(resp.RawBody())
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			common.ConversationDoneChannel <- true
			break
		}

		makeConversationResponse := parseEvent(line)
		if makeConversationResponse != nil {
			parts := makeConversationResponse.Message.Content.Parts
			if len(parts) != 0 {
				common.ResponseTextChannel <- parts[0]
			}
		}
	}

	if common.ConversationID == "" {
		go api.GenerateConversationTitle(tempConversationID)
	} else {
		common.ReloadConversationsChannel <- true
	}
}

func parseEvent(line string) *common.MakeConversationResponse {
	if strings.Contains(line, "DONE") {
		return nil
	}

	if strings.HasPrefix(line, "data: ") {
		var makeConversationResponse common.MakeConversationResponse
		str := strings.TrimRight(strings.TrimPrefix(line, "data: "), "\n")
		err := json.Unmarshal([]byte(str), &makeConversationResponse)
		if err != nil {
			return nil
		}

		if common.ParentMessageID == "" {
			common.ParentMessageID = makeConversationResponse.Message.ID
		}
		if common.ConversationID == "" {
			tempConversationID = makeConversationResponse.ConversationID
		}

		return &makeConversationResponse
	}

	return nil
}

func (api *API) GenerateConversationTitle(conversationID string) {
	_, err := client.R().
		SetBody(map[string]string{
			"message_id": common.MessageID,
		}).
		Post("/conversation/gen_title/" + conversationID)
	if err != nil {
		return
	}

	common.ReloadConversationsChannel <- true
}

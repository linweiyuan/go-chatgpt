package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/linweiyuan/go-chatgpt/common"
)

var (
	chatGPTClient *resty.Client

	apiClient *resty.Client
)

type API struct {
}

func New() *API {
	return &API{}
}

func init() {
	serverUrl := os.Getenv("SERVER_URL")
	if serverUrl == "" {
		log.Fatal("Please set server url first.")
	}
	accessToken := os.Getenv("ACCESS_TOKEN")
	if accessToken == "" {
		log.Fatal("Please set access token first.")
	}
	chatGPTClient = resty.New().SetBaseURL(serverUrl)
	chatGPTClient.SetHeader("Authorization", accessToken)

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("Please set api key first.")
	}
	apiClient = resty.New().SetBaseURL(serverUrl)
	apiClient.SetHeader("Authorization", apiKey)
}

//goland:noinspection GoUnhandledErrorResult
func (api *API) GetConversations() *common.Conversations {
	resp, _ := chatGPTClient.R().Get("/conversations?offset=0&limit=100")

	var conversations common.Conversations
	json.Unmarshal(resp.Body(), &conversations)

	return &conversations
}

//goland:noinspection GoUnhandledErrorResult
func (api *API) GetConversation(conversationID string) {
	resp, _ := chatGPTClient.R().Get("/conversation/" + conversationID)

	var conversation common.Conversation
	json.Unmarshal(resp.Body(), &conversation)

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
		if message.Author.Role == common.RoleUser {
			common.MessageChannel <- message
		}
	}
}

var tempConversationID string

//goland:noinspection GoUnhandledErrorResult
func (api *API) StartConversation(content string) {
	parentMessageID := common.ParentMessageID
	if parentMessageID == "" || common.ConversationID == "" {
		parentMessageID = uuid.NewString()
	}
	resp, _ := chatGPTClient.R().
		SetDoNotParseResponse(true).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "text/event-stream").
		SetBody(fmt.Sprintf(`
		{
			"action": "next",
			"messages": [{
				"id": "%s",
				"author": {
					"role": "%s"
				},
				"role": "%s",
				"content": {
					"content_type": "text",
					"parts": ["%s"]
				}
			}],
			"parent_message_id": "%s",
			"model": "%s",
			"conversation_id": "%s",
			"timezone_offset_min": -480,
			"variant_purpose": "none",
			"continue_text": "continue"
		}`, uuid.NewString(), common.RoleUser, common.RoleUser, content, parentMessageID, common.ChatGPTModel, common.ConversationID)).Post("/conversation")

	// get it again from response
	common.ParentMessageID = ""

	defer func(body io.ReadCloser) {
		body.Close()
	}(resp.RawBody())

	reader := bufio.NewReader(resp.RawBody())
	for {
		line, err := reader.ReadString('\n')
		if line == "\n" {
			continue
		}

		if strings.HasSuffix(line, "[DONE]\n") || err != nil {
			common.ConversationDoneChannel <- true
			break
		}

		var makeConversationResponse *common.StartConversationResponse
		json.Unmarshal([]byte(line[5:]), &makeConversationResponse)

		if common.ParentMessageID == "" {
			common.ParentMessageID = makeConversationResponse.Message.ID
		}
		if common.ConversationID == "" && tempConversationID == "" {
			tempConversationID = makeConversationResponse.ConversationID
		}

		if makeConversationResponse != nil {
			parts := makeConversationResponse.Message.Content.Parts
			if len(parts) != 0 {
				common.ResponseTextChannel <- parts[0]
			}
			if makeConversationResponse.Message.EndTurn == true {
				common.ConversationDoneChannel <- true
				break
			}
		}
	}

	if common.ConversationID == "" {
		go api.GenerateTitle(tempConversationID)
	} else {
		common.ReloadConversationsChannel <- true
	}
}

func (api *API) GenerateTitle(conversationID string) {
	_, err := chatGPTClient.R().
		SetBody(map[string]string{
			"message_id": common.ParentMessageID,
			"model":      common.ChatGPTModel,
		}).
		Post("/conversation/gen_title/" + conversationID)
	if err != nil {
		return
	}

	common.ReloadConversationsChannel <- true
}

func (api *API) RenameTitle(conversationID string, title string) {
	_, err := chatGPTClient.R().
		SetBody(map[string]string{
			"title": title,
		}).
		Patch("/conversation/" + conversationID)
	if err != nil {
		return
	}

	// seems no need to reload conversation list
}

func (api *API) DeleteConversation(conversationID string) {
	_, err := chatGPTClient.R().
		SetBody(map[string]bool{
			"is_visible": false,
		}).
		Patch("/conversation/" + conversationID)
	if err != nil {
		return
	}

	common.ReloadConversationsChannel <- true
}

func (api *API) ClearConversations() {
	_, err := chatGPTClient.R().
		SetBody(map[string]bool{
			"is_visible": false,
		}).
		Patch("/conversations")
	if err != nil {
		return
	}

	common.ConversationID = ""
	common.CurrentNode = nil
	common.ReloadConversationsChannel <- true
}

//goland:noinspection GoUnhandledErrorResult
func (api *API) ChatCompletions(content string) {
	chatCompletionsMessage := common.ChatCompletionsMessage{
		Role:    common.RoleUser,
		Content: content,
	}
	common.ApiMessages = append(common.ApiMessages, chatCompletionsMessage)
	chatCompletionsRequest := common.ChatCompletionsRequest{
		Model:    common.ApiModel,
		Messages: common.ApiMessages,
		Stream:   true,
	}
	data, _ := json.Marshal(chatCompletionsRequest)
	resp, _ := apiClient.R().
		SetDoNotParseResponse(true).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "text/event-stream").
		SetBody(data).
		Post(common.ApiVersion + "/chat/completions")
	defer func(body io.ReadCloser) {
		body.Close()
	}(resp.RawBody())

	responseContent := ""
	reader := bufio.NewReader(resp.RawBody())
	for {
		line, err := reader.ReadString('\n')
		if line == "\n" {
			continue
		}

		if strings.HasSuffix(line, "[DONE]\n") || err != nil {
			common.ConversationDoneChannel <- true
			break
		}

		var response *common.ChatCompletionsResponse
		json.Unmarshal([]byte(line[5:]), &response)

		if response != nil {
			choice := response.Choices[0]
			content := choice.Delta.Content
			if content != "" {
				responseContent += content
				common.ResponseTextChannel <- content
			}
			if choice.FinishReason == "stop" {
				common.ConversationDoneChannel <- true
				break
			}
		}
	}
	common.ApiMessages = append(common.ApiMessages, common.ChatCompletionsMessage{
		Role:    common.RoleAssistant,
		Content: responseContent,
	})
}

//goland:noinspection GoUnhandledErrorResult
func (api *API) CheckUsage() *common.CheckUsageResponse {
	resp, _ := apiClient.R().Get("/dashboard/billing/credit_grants")

	var checkUsageResponse common.CheckUsageResponse
	json.Unmarshal(resp.Body(), &checkUsageResponse)

	return &checkUsageResponse
}

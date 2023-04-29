package gpt

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/assistant-ai/llmchat-client/db"
)

type GptClient struct {
	openAiKey      string
	contextDepth   int
	model          GPTModel
	defaultContext string
}

func NewDefaultGptClient(openAiKey string) *GptClient {
	return &GptClient{
		openAiKey:      openAiKey,
		contextDepth:   5,
		model:          ModelGPT4,
		defaultContext: "",
	}
}

func NewGptClient(openAiKey string, contextDepth int, model GPTModel, defaultContext string) *GptClient {
	return &GptClient{
		openAiKey:      openAiKey,
		contextDepth:   contextDepth,
		model:          model,
		defaultContext: defaultContext,
	}
}

func NewGptClientFromFile(openAiKeyFlePath string, contextDepth int, model GPTModel, defaultContext string) (*GptClient, error) {
	b, err := os.ReadFile(openAiKeyFlePath)
	if err != nil {
		return nil, err
	}
	return &GptClient{
		openAiKey:      strings.ReplaceAll(string(b), "\n", ""),
		contextDepth:   contextDepth,
		model:          model,
		defaultContext: defaultContext,
	}, nil
}

func (g *GptClient) SendMessage(message string, inputContextId string) (string, error) {
	messages := make([]db.Message, 0)
	contextId := inputContextId
	if contextId == "" {
		contextId = db.RandomContextId
	}
	messages, err := db.GetMessagesByContextID(contextId)
	if len(messages) > g.contextDepth {
		messages = messages[len(messages)-g.contextDepth:]
	}
	if err != nil {
		return "", err
	}
	if g.defaultContext != "" {
		defaultContextMessage := db.CreateNewMessage(db.SystemRoleName, g.defaultContext, contextId)
		messages = append([]db.Message{defaultContextMessage}, messages...)
	}
	newMessage := db.CreateNewMessage(db.UserRoleName, message, contextId)
	_, err = db.StoreMessage(newMessage)
	if err != nil {
		return "", err
	}
	messages = append(messages, newMessage)
	contextMessage, err := db.GetContextMessage(contextId)
	if err != nil {
		return "", err
	}
	messages = append([]db.Message{contextMessage}, messages...)
	answers, err := g.sendMessages(messages, contextId)
	if err != nil {
		return "", err
	}
	return answers[len(answers)-1].Content, nil
}

func (g *GptClient) sendMessages(messages []db.Message, contextId string) ([]db.Message, error) {
	requestBody, err := g.prepareGPTRequestBody(messages)
	if err != nil {
		return nil, err
	}

	response, err := g.sendGPTRequest(requestBody)
	if err != nil {
		return nil, err
	}

	return addGPTResponse(response, messages, contextId)
}

func addGPTResponse(response *GptChatCompletionMessage, messages []db.Message, contextId string) ([]db.Message, error) {
	gpt4Text := response.Choices[0].Message.Content
	newMessage := db.CreateNewMessage(db.AssistentRoleNeam, gpt4Text, contextId)
	newMessage.ContextId = contextId
	messages = append(messages, newMessage)

	return messages, nil
}

func (g *GptClient) sendGPTRequest(requestBody []byte) (*GptChatCompletionMessage, error) {
	req, err := http.NewRequest("POST", API_URL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", g.openAiKey))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response GptChatCompletionMessage
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	if len(response.Choices) == 0 {
		return nil, errors.New("Error response from GPT")
	}

	return &response, nil
}

func (g *GptClient) prepareGPTRequestBody(messages []db.Message) ([]byte, error) {
	gptMessages := convertMessagesToMaps(messages)

	requestBody, err := json.Marshal(map[string]interface{}{
		"messages":   gptMessages,
		"max_tokens": 2000,
		"n":          1,
		"model":      g.model,
	})

	if err != nil {
		return nil, err
	}

	return requestBody, nil
}

func convertMessagesToMaps(messages []db.Message) []map[string]string {
	gptMessages := make([]map[string]string, len(messages))

	for i, message := range messages {
		formattedTimestamp := message.Timestamp.Format("2006-01-02 15:04:05")
		combinedContent := fmt.Sprintf("%s: %s", formattedTimestamp, message.Content)

		gptMessages[i] = map[string]string{
			"role":    message.Role,
			"content": combinedContent,
		}
	}

	return gptMessages
}

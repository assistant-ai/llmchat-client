package gpt

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/assistant-ai/llmchat-client/db"
)

type GptClient struct {
	OpenAiKey      string
	ContextDepth   int
	Model          *GPTModel
	DefaultContext string
	MaxTokens      int
}

func NewDefaultGptClient(openAiKey string) *GptClient {
	return &GptClient{
		OpenAiKey:      openAiKey,
		ContextDepth:   5,
		Model:          ModelGPT4,
		DefaultContext: "",
		MaxTokens:      8000,
	}
}

func NewDefaultGptClientFromFile(openAiKeyFilePath string) (*GptClient, error) {
	b, err := os.ReadFile(openAiKeyFilePath)
	if err != nil {
		return nil, err
	}
	return NewDefaultGptClient(strings.ReplaceAll(string(b), "\n", "")), nil
}

func NewGptClient(openAiKey string, contextDepth int, model *GPTModel, defaultContext string, maxTokens int) *GptClient {
	return &GptClient{
		OpenAiKey:      openAiKey,
		ContextDepth:   contextDepth,
		Model:          model,
		DefaultContext: defaultContext,
		MaxTokens:      maxTokens,
	}
}

func NewGptClientFromFile(openAiKeyFlePath string, contextDepth int, model *GPTModel, defaultContext string, maxTokens int) (*GptClient, error) {
	b, err := os.ReadFile(openAiKeyFlePath)
	if err != nil {
		return nil, err
	}
	return &GptClient{
		OpenAiKey:      strings.ReplaceAll(string(b), "\n", ""),
		ContextDepth:   contextDepth,
		Model:          model,
		DefaultContext: defaultContext,
		MaxTokens:      maxTokens,
	}, nil
}

func (g *GptClient) SenRandomContextMessage(message string) (string, error) {
	return g.SendMessage(message, db.RandomContextId)
}

func (g *GptClient) SendMessage(message string, inputContextId string) (string, error) {
	return g.SendMessageWithContextDepth(message, inputContextId, g.ContextDepth, true)
}

func (g *GptClient) SendMessageWithContextDepth(message string, inputContextId string, contextDepth int, addAllSystemContext bool) (string, error) {
	messages := make([]db.Message, 0)
	contextId := inputContextId
	if contextId == "" || contextId == db.RandomContextId {
		contextId = db.RandomContextId
	} else {
		count := contextDepth
		messagesFromDb, err := db.GetLastMessagesByContextID(contextId, count)
		if err != nil {
			return "", err
		}
		messages = messagesFromDb
	}
	if addAllSystemContext {
		if g.DefaultContext != "" {
			defaultContextMessage := db.CreateNewMessage(db.SystemRoleName, g.DefaultContext, contextId)
			messages = append([]db.Message{defaultContextMessage}, messages...)
		}
		userDefaultContextExist, err := db.CheckIfUserDefaultContextExists()
		if err != nil {
			return "", err
		}
		if userDefaultContextExist {
			userDefaultContextMessage, err := db.GetUserDefaultContextMessage()
			if err != nil {
				return "", err
			}
			messages = append([]db.Message{userDefaultContextMessage}, messages...)
		}
	}
	newMessage := db.CreateNewMessage(db.UserRoleName, message, contextId)
	_, err := db.StoreMessage(newMessage)
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
	answerMessage := answers[len(answers)-1]
	_, err = db.StoreMessage(answerMessage)
	if err != nil {
		return "", err
	}
	return answerMessage.Content, nil
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

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", g.OpenAiKey))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response GptChatCompletionMessage
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var decodedString = string(bodyBytes)
	reader := strings.NewReader(decodedString)
	if err := json.NewDecoder(reader).Decode(&response); err != nil {
		return nil, err
	}

	if len(response.Choices) == 0 {
		return nil, errors.New(fmt.Sprintf("Error response from GPT: %s", decodedString))
	}

	return &response, nil
}

func sumOfTokensAcrossAllMessages(messages []map[string]string) int {
	tokens := 0
	for _, message := range messages {
		tokens += len(message["content"])
		tokens += len(message["role"])
	}
	return tokens / 3
}

func (g *GptClient) prepareGPTRequestBody(messages []db.Message) ([]byte, error) {
	gptMessages := convertMessagesToMaps(messages)
	tokens := sumOfTokensAcrossAllMessages(gptMessages)
	maxTokens := g.MaxTokens
	if tokens+g.MaxTokens >= g.Model.MaxTokens {
		maxTokens = g.MaxTokens - tokens
	}

	if maxTokens <= 0 {
		return nil, errors.New("Not enough tokens")
	}

	requestBody, err := json.Marshal(map[string]interface{}{
		"messages":   gptMessages,
		"max_tokens": maxTokens,
		"n":          1,
		"model":      g.Model.Name,
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

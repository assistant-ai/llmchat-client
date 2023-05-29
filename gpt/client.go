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

	"github.com/assistant-ai/llmchat-client/client"
	"github.com/assistant-ai/llmchat-client/db"
)

type GptClient struct {
	OpenAiKey string
	Model     *GPTModel
	MaxTokens int
}

func NewDefaultGptClient(openAiKey string) *client.Client {
	return &client.Client{
		Client: &GptClient{
			OpenAiKey: openAiKey,
			Model:     ModelGPT4,
			MaxTokens: 8000,
		},
		ContextDepth:   5,
		DefaultContext: "",
	}
}

func NewDefaultGptClientFromFile(openAiKeyFilePath string) (*client.Client, error) {
	b, err := os.ReadFile(openAiKeyFilePath)
	if err != nil {
		return nil, err
	}
	return NewDefaultGptClient(strings.ReplaceAll(string(b), "\n", "")), nil
}

func NewGptClient(openAiKey string, contextDepth int, model *GPTModel, defaultContext string, maxTokens int) *client.Client {
	return &client.Client{
		Client: &GptClient{
			OpenAiKey: openAiKey,
			Model:     ModelGPT4,
			MaxTokens: maxTokens,
		},
		ContextDepth:   contextDepth,
		DefaultContext: defaultContext,
	}
}

func NewGptClientFromFile(openAiKeyFlePath string, contextDepth int, model *GPTModel, defaultContext string, maxTokens int) (*client.Client, error) {
	b, err := os.ReadFile(openAiKeyFlePath)
	if err != nil {
		return nil, err
	}
	return &client.Client{
		Client: &GptClient{
			OpenAiKey: strings.ReplaceAll(string(b), "\n", ""),
			Model:     ModelGPT4,
			MaxTokens: maxTokens,
		},
		ContextDepth:   contextDepth,
		DefaultContext: defaultContext,
	}, nil
}

func (g *GptClient) SendMessages(messages []db.Message, context []string) ([]db.Message, error) {
	contextId := messages[0].ContextId
	for _, contextMsg := range context {
		messages = append(messages, db.CreateNewMessage(db.SystemRoleName, contextMsg, contextId))
	}
	requestBody, err := g.prepareGPTRequestBody(messages)
	if err != nil {
		return nil, err
	}

	response, err := g.sendGPTRequest(requestBody)
	if err != nil {
		return nil, err
	}

	return addGPTResponse(response, messages)
}

func addGPTResponse(response *GptChatCompletionMessage, messages []db.Message) ([]db.Message, error) {
	gpt4Text := response.Choices[0].Message.Content
	newMessage := db.CreateNewMessage(db.AssistentRoleNeam, gpt4Text, messages[0].ContextId)
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
	model := g.Model
	if tokens+g.MaxTokens >= g.Model.MaxTokens {
		maxTokens = model.MaxTokens - tokens
		if g.Model == ModelGPT4 && tokens > g.Model.MaxTokens/2 {
			model = ModelGPT4Big
			maxTokens = model.MaxTokens - tokens
		}
	}

	if maxTokens <= 0 {
		return nil, errors.New("Not enough tokens")
	}

	requestBody, err := json.Marshal(map[string]interface{}{
		"messages":   gptMessages,
		"max_tokens": maxTokens,
		"n":          1,
		"model":      model.Name,
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

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
	"github.com/sirupsen/logrus"
)

type GptClient struct {
	OpenAiKey string
	Model     *GPTModel
	MaxTokens int
	Logger    *logrus.Logger
}

func NewDefaultGptClient(openAiKey string, logger *logrus.Logger) *client.Client {
	return NewGptClient(openAiKey, 5, ModelGPT4, "", 8000, logger)
}

func NewDefaultGptClientFromFile(openAiKeyFilePath string, logger *logrus.Logger) (*client.Client, error) {
	b, err := os.ReadFile(openAiKeyFilePath)
	if err != nil {
		return nil, err
	}
	return NewDefaultGptClient(strings.ReplaceAll(string(b), "\n", ""), logger), nil
}

func NewGptClient(openAiKey string, contextDepth int, model *GPTModel, defaultContext string, maxTokens int, logger *logrus.Logger) *client.Client {
	return &client.Client{
		Client: &GptClient{
			OpenAiKey: openAiKey,
			Model:     model,
			MaxTokens: maxTokens,
			Logger:    logger,
		},
		ContextDepth:   contextDepth,
		DefaultContext: defaultContext,
		Logger:         logger,
	}
}

func NewGptClientFromFile(openAiKeyFlePath string, contextDepth int, model *GPTModel, defaultContext string, maxTokens int, logger *logrus.Logger) (*client.Client, error) {
	b, err := os.ReadFile(openAiKeyFlePath)
	if err != nil {
		return nil, err
	}
	return NewGptClient(strings.ReplaceAll(string(b), "\n", ""), contextDepth, model, defaultContext, maxTokens, logger), nil
}

func (g *GptClient) SendMessages(messages []db.Message, context []string) ([]db.Message, error) {
	contextId := messages[0].ContextId
	for _, contextMsg := range context {
		if g.Logger != nil {
			g.Logger.WithFields(logrus.Fields{
				"contextMsg": contextMsg,
			}).Debug("GPT context msg")
		}
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
		// Disabled since GPT4 big is not yet available
		// if g.Model == ModelGPT4 && tokens > g.Model.MaxTokens/2 {
		// 	model = ModelGPT4Big
		// 	maxTokens = model.MaxTokens - tokens
		// }
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

	if g.Logger != nil {
		g.Logger.WithFields(logrus.Fields{
			"requestBody": string(requestBody),
		}).Debug("GPT request")
	}

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

package palm
import (
	"fmt"
	"context"
	"encoding/json"
	"net/http"
	"bytes"
	"io/ioutil"
	"github.com/assistant-ai/llmchat-client/db"
	"github.com/assistant-ai/llmchat-client/client"
	"golang.org/x/oauth2/google"
)

type PalmClient struct {
	GCPAccessToken string
	GCPProjectId string
}

func NewDefaultTokenPalmClient(GCPProjectId string) (*client.Client, error) {
	token, err := getDefaultAccesstToken()
	if (err != nil) {
		return nil, err
	}
	return &client.Client{
		Client:  &PalmClient{
			GCPAccessToken: token,
			GCPProjectId: GCPProjectId,
		},
		ContextDepth:   8,
		DefaultContext: db.RandomContextId,
	}, nil
}

func NewPalmClient(GCPProjectId string, serviceAccountJsonPath string) (*client.Client, error) {
	token, err := getAccessTokenFromFile(serviceAccountJsonPath)
	if (err != nil) {
		return nil, err
	}
	return &client.Client{
		Client:  &PalmClient{
			GCPAccessToken: token,
			GCPProjectId: GCPProjectId,
		},
		ContextDepth:   8,
		DefaultContext: db.RandomContextId,
	}, nil
}

func getDefaultAccesstToken() (string, error) {
	ctx := context.Background()
	creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return "", fmt.Errorf("unable to find default credentials: %v", err)
	}
	return getAccessToken(creds)
}

func getAccessTokenFromFile(saFilePath string) (string, error) {
	ctx := context.Background()

	// Read service account file content
	jsonKey, err := ioutil.ReadFile(saFilePath)
	if err != nil {
		return "", fmt.Errorf("unable to read service account file: %v", err)
	}

	// Parse the credentials from the JSON key
	creds, err := google.CredentialsFromJSON(ctx, jsonKey, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return "", fmt.Errorf("unable to parse service account credentials: %v", err)
	}
	return getAccessToken(creds)
}

func getAccessToken(creds *google.Credentials) (string, error) {
	tokenSource := creds.TokenSource
	token, err := tokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("unable to obtain access token: %v", err)
	}

	return token.AccessToken, nil
}

func hackyZipAllMessagesInOne(messages []PalmMessage) []PalmMessage {
	finalMessage := ""
	for _, msg := range messages {
		finalMessage = fmt.Sprintf("%s: %s\n", msg.Author, msg.Content)
	}
	newMessages := make([]PalmMessage, 1)
	newMessages[0] = PalmMessage{
		Author: "user",
		Content: finalMessage,
	}
	return newMessages
}

func (c *PalmClient) SendMessages(messages []db.Message, context []string) ([]db.Message, error) {
	apiEndpoint := "us-central1-aiplatform.googleapis.com"
	modelID := "chat-bison"

	url := fmt.Sprintf("https://%s/v1/projects/%s/locations/us-central1/publishers/google/models/%s:predict", apiEndpoint, c.GCPProjectId, modelID)

	palmMessages := make([]PalmMessage, 0, len(messages))
	for _, msg := range messages {
		palmMessages = append(palmMessages, dbMessageToPalmMessage(msg))
	}

	finalContext := ""

	for _, contextMsg := range context {
		finalContext = finalContext + "\n" + contextMsg
	}

	palmMessages, err := TrimIfNeeded(palmMessages, finalContext)
	if err != nil {
		return nil, err
	}

	palmInstance := &PalmInstance{
		Context: finalContext,
		Examples: make([]string, 0),
		Messages: hackyZipAllMessagesInOne(palmMessages),
	}
	palmInstances := make([]PalmInstance, 1)
	palmInstances[0] = *palmInstance

	payload := PredictPayload{
		Instances: palmInstances,
		Parameters: Parameters{
			Temperature:     0.2,
			MaxOutputTokens: 1000,
			TopP:            0.9,
			TopK:            40,
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.GCPAccessToken))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var predictResp PredictResponse
	
	err = json.Unmarshal(responseBody, &predictResp)
	if err != nil {
		return nil, err
	}

	// Extract and append response messages (PalmMessage format)
	newDbMessages := make([]db.Message, 0, len(predictResp.Predictions))
	for _, prediction := range predictResp.Predictions {
		for _, candidate := range prediction.Candidates {
			newDbMessages = append(newDbMessages, db.CreateNewMessage(db.AssistentRoleNeam, candidate.Content, messages[0].ContextId))
		}
	}

	messages = append(messages, newDbMessages...)

	// Return the modified messages slice (including responses)
	return messages, nil
}

func dbMessageToPalmMessage(dbMsg db.Message) PalmMessage {
	return PalmMessage{
		Author:  dbMsg.Role,
		Content: dbMsg.Content,
	}
}

package gpt

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSendMessage(t *testing.T) {
	// Replace the API key with your OpenAI API key
	apiKeyFilePath := filepath.Join(os.Getenv("HOME"), ".open-ai.key")
	contextDepth := 5
	client, err := NewGptClientFromFile(apiKeyFilePath, contextDepth, ModelGPT4, "", 8000)

	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// Test input and expected response
	testMessage := "Hello, AI assistant! How are you?"
	inputContextId := "12"

	response, err := client.SendMessage(testMessage, inputContextId)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	t.Log("Assistant response:", response)

	if response == "" {
		t.Fatalf("Assistant response is empty")
	}
}

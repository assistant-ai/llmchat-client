package llama2_client

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSendMessage(t *testing.T) {
	// Replace the API key with your OpenAI API key
	checkpointPath := filepath.Join(os.Getenv("HOME"), "llama/llama-2-7b-chat.bin")
	tokenizerPath := filepath.Join(os.Getenv("HOME"), "llama/tokenizer2.bin")
	temperature := 0.9
	steps := 256
	topp := 0.9
	contextDepth := 5
	context := "test"
	client := NewLlama2Client(checkpointPath, tokenizerPath, temperature, steps, topp, nil, contextDepth, context)

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

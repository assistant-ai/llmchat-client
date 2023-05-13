package palm

import (
	"testing"
)

func TestSendMessage(t *testing.T) {
	client, err := NewPalmClient("ml-lab-152505", "/Users/slava.kovalevskyi/Downloads/ml-lab-152505-1829490d6dd7.json")

	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// Test input and expected response
	testMessage := "Hello, AI assistant! How are you?"
	inputContextId := "palm"

	response, err := client.SendMessage(testMessage, inputContextId)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	t.Log("Assistant response:", response)

	if response == "" {
		t.Fatalf("Assistant response is empty")
	}
}

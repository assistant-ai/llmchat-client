package palm

import (
	"fmt"
)

func countTokens(messages []PalmMessage, context string) int {
		totalInputTokens := 0
		for _, msg := range messages {
			totalInputTokens += len(msg.Content)
		}
		totalInputTokens += len(context)
		return totalInputTokens
}

func TrimIfNeeded(messages []PalmMessage, context string) ([]PalmMessage, error) {
	totalInputTokens := countTokens(messages, context)
	for totalInputTokens >= MAX_INPUT_TOKENS && len(messages) > 0 {
		// Return an error if even one message + context is still bigger
		if len(messages) == 1 {
			return nil, fmt.Errorf("the message and context size (%d tokens) exceeds the maximum allowed input tokens (%d)", totalInputTokens, MAX_INPUT_TOKENS)
		}
		messages = messages[1:]
		totalInputTokens = countTokens(messages, context)
	}	
	return messages, nil
}

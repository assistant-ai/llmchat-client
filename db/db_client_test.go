package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetContextMessage(t *testing.T) {
	// Set up test context
	contextID := "testContextID"
	expectedContext := "Test context content"
	err := CreateContext(contextID, expectedContext)
	if err != nil {
		t.Errorf("Failed to create test context: %v", err)
		return
	}

	defer func() {
		// Clean up test context
		err := RemoveContext(contextID)
		if err != nil {
			t.Errorf("Failed to clean up test context: %v", err)
		}
	}()

	// Call GetContextMessage function
	message, err := GetContextMessage(contextID)
	if err != nil {
		t.Errorf("GetContextMessage failed: %v", err)
		return
	}

	// Assert context content
	assert.Equal(t, contextID, message.ContextId, "Returned message has wrong context ID")
	assert.Equal(t, SystemRoleName, message.Role, "Returned message has wrong role")
	assert.Equal(t, expectedContext, message.Content, "Returned message has wrong content")
}

package db

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID        string    `json:"id"`
	ContextId string    `json:"context_id"`
	Timestamp time.Time `json:"timestamp"`
	Role      string    `json:"sender"`
	Content   string    `json:"content"`
}

func CreateNewMessage(role string, content string, contextId string) Message {
	uuidMsg, _ := uuid.NewUUID()
	idMsg := uuidMsg.String()

	return Message{
		ID:        idMsg,
		ContextId: contextId,
		Timestamp: time.Now(),
		Role:      role,
		Content:   content,
	}
}

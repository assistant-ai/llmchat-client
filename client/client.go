package client

import (
	"github.com/assistant-ai/llmchat-client/db"
)

type Client struct {
	Client         LllmChatClient
	ContextDepth   int
	DefaultContext string
}

type LllmChatClient interface {
	SendMessages([]db.Message) ([]db.Message, error)
}

func (c *Client) SenRandomContextMessage(message string) (string, error) {
	return c.SendMessage(message, db.RandomContextId)
}

func (c *Client) SendMessage(message string, inputContextId string) (string, error) {
	return c.SendMessageWithContextDepth(message, inputContextId, c.ContextDepth, true)
}

func (c *Client) SendMessageWithContextDepth(message string, inputContextId string, contextDepth int, addAllSystemContext bool) (string, error) {
	messages := make([]db.Message, 0)
	contextId := inputContextId

	existContext, err := db.CheckIfContextExists(contextId)
	if err != nil {
		return "", err
	}
	if !existContext {
		db.CreateContext(contextId, "")
	}

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
		if c.DefaultContext != "" {
			defaultContextMessage := db.CreateNewMessage(db.SystemRoleName, c.DefaultContext, contextId)
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
		contextMessage, err := db.GetContextMessage(contextId)
		if err != nil {
			return "", err
		}
		messages = append([]db.Message{contextMessage}, messages...)
	}
	newMessage := db.CreateNewMessage(db.UserRoleName, message, contextId)
	_, err = db.StoreMessage(newMessage)
	if err != nil {
		return "", err
	}
	messages = append(messages, newMessage)
	answers, err := (c.Client).SendMessages(messages)
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
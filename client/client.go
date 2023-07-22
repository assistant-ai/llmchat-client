package client

import (
	"github.com/assistant-ai/llmchat-client/db"
	"github.com/sirupsen/logrus"
)

type Client struct {
	Client         LllmChatClient
	ContextDepth   int
	DefaultContext string
	Logger         *logrus.Logger
}

type LllmChatClient interface {
	SendMessages(messages []db.Message, context []string) ([]db.Message, error)
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
	if c.Logger != nil {
		c.Logger.WithFields(logrus.Fields{
			"message":           message,
			"contextId":         contextId,
			"contextDepth":      contextDepth,
			"addAllSystemConte": addAllSystemContext,
		}).Debug("Send message")
	}
	context := make([]string, 0)
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
			context = append(context, c.DefaultContext)
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
			context = append(context, userDefaultContextMessage)
		}
	}

	contextExist, err := db.CheckIfContextExists(contextId)
	if err != nil {
		return "", err
	}

	contextMessage := ""
	if contextExist {
		contextMessage, err := db.GetContextMessage(contextId)
		if err != nil {
			return "", err
		}
		if c.Logger != nil {
			c.Logger.WithFields(logrus.Fields{
				"contextMessage": contextMessage,
				"contextId":      contextId,
			}).Debug("Context message")
		}
	}

	if contextMessage != "" {
		context = append(context, contextMessage)
	}
	newMessage := db.CreateNewMessage(db.UserRoleName, message, contextId)
	_, err = db.StoreMessage(newMessage)
	if err != nil {
		return "", err
	}
	messages = append(messages, newMessage)
	answers, err := (c.Client).SendMessages(messages, context)
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

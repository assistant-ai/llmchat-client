package db

import (
	"database/sql"
	"log"
	"path/filepath"

	"github.com/google/uuid"

	"github.com/b0noi/go-utils/v2/fs"
	_ "github.com/mattn/go-sqlite3"
)

const DefaultContextID = "defaultUserContext"

var db *sql.DB

func init() {
	var err error
	folderPath, err := fs.MaybeCreateProgramFolder("llmchat-client")
	if err != nil {
		log.Fatal(err)
	}
	dbFilePath := filepath.Join(folderPath, "messages.db")
	db, err = sql.Open("sqlite3", dbFilePath)
	if err != nil {
		log.Fatal(err)
	}

	createMessagesTable := `CREATE TABLE IF NOT EXISTS messages (
		id TEXT PRIMARY KEY,
		context_id TEXT,
		timestamp DATETIME,
		role TEXT,
		content TEXT
	);`

	_, err = db.Exec(createMessagesTable)
	if err != nil {
		log.Fatal(err)
	}

	createContextTable := `CREATE TABLE IF NOT EXISTS context (
		context_id TEXT PRIMARY KEY,
		context TEXT
	);`

	_, err = db.Exec(createContextTable)
	if err != nil {
		log.Fatal(err)
	}
}

func RemoveContext(contextId string) error {
	if err := removeById(`DELETE FROM messages WHERE context_id = ?`, contextId); err != nil {
		return err
	}
	if err := removeById(`DELETE FROM context WHERE context_id = ?`, contextId); err != nil {
		return err
	}
	return nil
}

func removeById(query string, id string) error {
	statement := query

	result, err := db.Exec(statement, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	log.Printf("Removed %d objects with id %s", rowsAffected, id)
	return nil
}

func CheckIfContextExists(contextId string) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM context WHERE context_id=? LIMIT 1)", contextId).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func CheckIfUserDefaultContextExists() (bool, error) {
	return CheckIfContextExists(DefaultContextID)
}

func GetUserDefaultContextMessage() (Message, error) {
	return GetContextMessage(DefaultContextID)
}

func UpdateUserDefaultContext(context string) error {
	return UpdateContext(DefaultContextID, context)
}

func UpdateContext(contextId string, context string) error {
	exist, err := CheckIfContextExists(contextId)
	if err != nil {
		return err
	}
	if !exist {
		return CreateContext(contextId, context)
	}
	stmt, err := db.Prepare("UPDATE context SET context=? WHERE context_id=?")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(context, contextId)
	if err != nil {
		return err
	}

	return nil
}

func CreateContext(contextId string, context string) error {
	stmt, err := db.Prepare("INSERT INTO context(context_id, context) VALUES(?, ?)")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(contextId, context)
	if err != nil {
		return err
	}

	return nil
}

func StoreMessage(m Message) (string, error) {
	context := m.ContextId
	contextExists, err := CheckIfContextExists(context)
	if err != nil {
		return "", err
	}
	if !contextExists {
		if err := CreateContext(context, ""); err != nil {
			return "", err
		}
	}
	if m.ID == "" {
		m.ID = uuid.New().String()
	}

	stmt, err := db.Prepare("INSERT INTO messages(id, context_id, timestamp, role, content) VALUES(?, ?, ?, ?, ?)")
	if err != nil {
		return "", err
	}

	_, err = stmt.Exec(m.ID, m.ContextId, m.Timestamp, m.Role, m.Content)
	if err != nil {
		return "", err
	}

	return m.ID, nil
}

func GetContextIDs() ([]string, error) {
	rows, err := db.Query("SELECT context_id FROM context")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	contextIDs := []string{}
	for rows.Next() {
		var contextID string
		err := rows.Scan(&contextID)
		if err != nil {
			return nil, err
		}
		contextIDs = append(contextIDs, contextID)
	}

	return contextIDs, nil
}

func GetMessageByID(id string) (Message, error) {
	var m Message
	err := db.QueryRow("SELECT id, context_id, timestamp, role, content FROM messages WHERE id=?", id).Scan(&m.ID, &m.ContextId, &m.Timestamp, &m.Role, &m.Content)
	if err != nil {
		return Message{}, err
	}

	return m, nil
}

func GetContextMessage(contextId string) (Message, error) {
	var m Message
	m.Role = SystemRoleName
	err := db.QueryRow("SELECT context_id, context FROM context WHERE context_id=?", contextId).Scan(&m.ContextId, &m.Content)
	if err != nil {
		return Message{}, err
	}

	return m, nil
}

func DeleteMessageByID(id string) error {
	stmt, err := db.Prepare("DELETE FROM messages WHERE id=?")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(id)
	if err != nil {
		return err
	}

	return nil
}

func GetMessagesByContextID(contextID string) ([]Message, error) {
	rows, err := db.Query("SELECT id, context_id, timestamp, role, content FROM messages WHERE context_id=? ORDER BY timestamp ASC", contextID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := []Message{}
	for rows.Next() {
		var m Message
		err := rows.Scan(&m.ID, &m.ContextId, &m.Timestamp, &m.Role, &m.Content)
		if err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}

	return messages, nil
}

package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

type DirectMessage struct {
	//TODO: maybe add message id or not
	// messageId uuid.UUID
	FromId  string    `json:"from_id"`
	ToId    string    `json:"to_id"`
	Body    string    `json:"body"`
	Created time.Time `json:"created"`
}

type DirectMessageModel struct {
	DB *sql.DB
}

func (m *DirectMessageModel) GetMessagesForUser(currentUserId, userId string) ([]*DirectMessage, error) {
	stmt := "SELECT body, created FROM direct_message WHERE from_id = $1 AND to_id = $2;"
	rows, err := m.DB.Query(stmt, currentUserId, userId)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	messages := []*DirectMessage{}

	for rows.Next() {
		msg := &DirectMessage{}
		err := rows.Scan(&msg.Body, &msg.Created)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return messages, nil
}
func (m *DirectMessageModel) Send(senderId, receiverId, msg string) error {
	stmt := "INSERT INTO direct_message (from_id, to_id, body, created) VALUES ($1,$2,$3,$4);"
	_, err := m.DB.Exec(stmt, senderId, receiverId, msg, time.Now().UTC())
	if err != nil {
		return err
	}
	return nil
}

func (m *DirectMessage) Serialize() ([]byte, error) {
	val, err := json.Marshal(m)
	if err != nil {
		//WARN: do something with this error
		return []byte{}, err
	}
	return val, nil
}

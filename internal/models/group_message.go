package models

import "time"

// should be the same for the most part, maybe changing to an interface instead of a concrete type  ?
type GroupMessage struct {
	//TODO: maybe add message id or not
	// messageId uuid.UUID
	FromId   string    `json:"from_id"` // userID
	ToId     string    `json:"to_id"`   // groupID
	Body     string    `json:"body"`
	Created  time.Time `json:"created"`
	Sender   string    `json:"sender"`   // not store in db
	Receiver string    `json:"receiver"` //not stored in db ; probably not used either
}

package domain

import (
	"encoding/json"
	"time"
)

type Message struct {
	ID           string    `json:"id,omitempty"`
	Type         string    `json:"type,omitempty"`
	Content      string    `json:"content"`
	RoomID       string    `json:"room_id"`
	UserID       string    `json:"user_id"`
	Username     string    `json:"username"`
	TargetUserID string    `json:"target_user_id,omitempty"`
	CreatedAt    time.Time `json:",omitempty"`
}

func FromJSON(payload []byte, msg *Message) error {
	return json.Unmarshal(payload, msg)
}

func (m *Message) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

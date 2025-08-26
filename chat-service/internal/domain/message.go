package domain

import "encoding/json"

type Message struct {
	Content  string `json:"content"`
	RoomID   string `json:"room_id"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

func FromJSON(payload []byte, msg *Message) error {
	return json.Unmarshal(payload, msg)
}

func (m *Message) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

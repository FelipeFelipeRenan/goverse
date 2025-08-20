package hub

type Message struct {
	Content string `json:"content"`
	RoomID string `json:"room_id"`
	UserID string `json:"user_id"`
	Username string `json:"username"`
}
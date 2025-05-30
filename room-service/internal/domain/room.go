package domain

import "time"

type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
	RoleGuest  Role = "guest"
)

type Room struct {
	ID          string    `json:"room_id"`
	Name        string    `json:"room_name"`
	Description string    `json:"room_description"`
	IsPublic    bool      `json:"is_public"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	OwnerID     string    `json:"owner_id"`
}

type RoomMember struct {
	RoomID   string    `json:"room_id"`
	UserID   string    `json:"user_id"`
	Role     Role      `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

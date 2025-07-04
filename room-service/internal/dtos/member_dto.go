package dtos

import (
	"time"

	"github.com/FelipeFelipeRenan/goverse/room-service/internal/domain"
)

type AddMemberRequest struct {
	UserID string      `json:"user_id"`
	Role   domain.Role `json:"role"`
}

type UpdateRoleRequest struct {
	NewRole domain.Role `json:"new_role"`
}

type UserDTO struct {
	ID        string `json:"user_id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Picture   string `json:"picture"`
	CreatedAt string `json:"created_at"`
	IsOAuth   bool   `json:"is_oauth"`
}

type MemberWithUser struct {
	RoomID   string    `json:"room_id"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
	User     UserDTO   `json:"user"`
}

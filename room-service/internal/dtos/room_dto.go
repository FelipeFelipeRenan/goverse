package dtos

import "github.com/FelipeFelipeRenan/goverse/room-service/internal/domain"

type CreateRoomRequest struct {
	Name        string `json:"name"`
	OwnerID     string `json:"owner_id"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
	MaxMembers  int    `json:"max_members"`
}

type RoomResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	OwnerID     string `json:"owner_id"`
	IsPublic    bool   `json:"is_private"`
	MemberCount int    `json:"member_count"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func FromRoom(room *domain.Room) *RoomResponse {
	return &RoomResponse{
		ID:          room.ID,
		Name:        room.Name,
		Description: room.Description,
		OwnerID:     room.OwnerID,
		IsPublic:    room.IsPublic,
		MemberCount: room.MemberCount,
		CreatedAt:   room.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   room.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

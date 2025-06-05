package dtos

import "github.com/FelipeFelipeRenan/goverse/room-service/internal/domain"

type AddMemberRequest struct {
	UserID string      `json:"user_id"`
	Role   domain.Role `json:"role"`
}

type UpdateRoleRequest struct {
	NewRole domain.Role `json:"new_role"`
}

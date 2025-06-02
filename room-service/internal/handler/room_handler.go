package handler

import (
	"context"

	"github.com/FelipeFelipeRenan/goverse/room-service/internal/client"
	"github.com/FelipeFelipeRenan/goverse/room-service/internal/domain"
	"github.com/FelipeFelipeRenan/goverse/room-service/internal/repository"
)

type MemberService interface {
	AddMember(ctx context.Context, actorID, roomID, userID string, role domain.Role) error
	RemoveMember(ctx context.Context, actorID, roomID, userID string) error
	UpdateMemberRole(ctx context.Context, actorID, roomID, userID string, newRole domain.Role) error
	IsUserValid(ctx context.Context, userID string) (bool, error)
}

type memberService struct {
	memberRepo repository.RoomMemberRepository
	roomRepo   repository.RoomRepository
	userClient client.UserServiceClient
}

func NewMemberService(m repository.RoomMemberRepository, r repository.RoomRepository, u client.UserServiceClient) MemberService {
	return &memberService{
		memberRepo: m,
		roomRepo:   r,
		userClient: u,
	}
}

// AddMember implements MemberService.
func (m *memberService) AddMember(ctx context.Context, actorID string, roomID string, userID string, role domain.Role) error {
	panic("unimplemented")
}

// IsUserValid implements MemberService.
func (m *memberService) IsUserValid(ctx context.Context, userID string) (bool, error) {
	panic("unimplemented")
}

// RemoveMember implements MemberService.
func (m *memberService) RemoveMember(ctx context.Context, actorID string, roomID string, userID string) error {
	panic("unimplemented")
}

// UpdateMemberRole implements MemberService.
func (m *memberService) UpdateMemberRole(ctx context.Context, actorID string, roomID string, userID string, newRole domain.Role) error {
	panic("unimplemented")
}

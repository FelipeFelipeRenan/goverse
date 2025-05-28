package service

import (
	"context"
	"time"

	"github.com/FelipeFelipeRenan/goverse/room-service/internal/domain"
	"github.com/FelipeFelipeRenan/goverse/room-service/internal/repository"
)

type RoomService interface {
	CreateRoom(ctx context.Context, ownerID string, room *domain.Room) (*domain.Room, error)
	DeleteRoom(ctx context.Context, actorID, roomID string) error
	AddMember(ctx context.Context, actorID, roomID, userID string, role domain.Role) error
	RemoveMember(ctx context.Context, actorID, roomID, userID string) error
	GetRoom(ctx context.Context, roomID string) (*domain.Room, error)
	GetRoomMembers(ctx context.Context, roomID string) ([]*domain.RoomMember, error)
}

type roomService struct {
	roomRepo   repository.RoomRepository
	memberRepo repository.RoomMemberRepository
}

func NewRoomService(r repository.RoomRepository, m repository.RoomMemberRepository) RoomService {
	return &roomService{
		roomRepo:   r,
		memberRepo: m,
	}
}

// CreateRoom implements RoomService.
func (r *roomService) CreateRoom(ctx context.Context, ownerID string, room *domain.Room) (*domain.Room, error) {
	room.OwnerID = ownerID
	room.CreatedAt = time.Now()

	err := r.roomRepo.Create(ctx, room)
	if err != nil {
		return nil, err
	}

	member := &domain.RoomMember{
		RoomID:   room.ID,
		UserID:   ownerID,
		Role:     domain.RoleOwner,
		JoinedAt: room.CreatedAt,
	}

	err = r.memberRepo.AddMember(ctx, member)
	if err != nil {
		_ = r.roomRepo.Delete(ctx, room.ID)
		return nil, err
	}
	return room, nil

}

// DeleteRoom implements RoomService.
func (r *roomService) DeleteRoom(ctx context.Context, actorID string, roomID string) error {
	panic("unimplemented")
}

// AddMember implements RoomService.
func (r *roomService) AddMember(ctx context.Context, actorID string, roomID string, userID string, role domain.Role) error {
	panic("unimplemented")
}

// GetRoom implements RoomService.
func (r *roomService) GetRoom(ctx context.Context, roomID string) (*domain.Room, error) {
	panic("unimplemented")
}

// GetRoomMembers implements RoomService.
func (r *roomService) GetRoomMembers(ctx context.Context, roomID string) ([]*domain.RoomMember, error) {
	panic("unimplemented")
}

// RemoveMember implements RoomService.
func (r *roomService) RemoveMember(ctx context.Context, actorID string, roomID string, userID string) error {
	panic("unimplemented")
}

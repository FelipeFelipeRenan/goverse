package service

import (
	"context"
	"fmt"
	"time"

	"github.com/FelipeFelipeRenan/goverse/room-service/internal/domain"
	"github.com/FelipeFelipeRenan/goverse/room-service/internal/repository"
)

type RoomService interface {
	CreateRoom(ctx context.Context, ownerID string, room *domain.Room) (*domain.Room, error)
	DeleteRoom(ctx context.Context, actorID, roomID string) error
	GetRoomByID(ctx context.Context, roomID string) (*domain.Room, error)
	ListRooms(ctx context.Context, limit, offset int, publicOnly bool, keyword string) ([]*domain.Room, error)
	UpdateRoom(ctx context.Context, actorID string, room *domain.Room) error
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
	// verifica se actor é o dono da sala (possui role de owner)
	room, err := r.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return fmt.Errorf("Sala nao encontrada: %w", err)
	}
	if room.OwnerID != actorID {
		return fmt.Errorf("somente o dono da sala pode deletá-la")
	}

	return r.roomRepo.Delete(ctx, roomID)
}

// GetRoom implements RoomService.
func (r *roomService) GetRoomByID(ctx context.Context, roomID string) (*domain.Room, error) {
	room, err := r.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	return room, nil
}

// GetRoomMembers implements RoomService.
func (r *roomService) GetRoomMembers(ctx context.Context, roomID string) ([]*domain.RoomMember, error) {
	return r.memberRepo.GetMembers(ctx, roomID)
}

// ListRooms implements RoomService.
func (r *roomService) ListRooms(ctx context.Context, limit int, offset int, publicOnly bool, keyword string) ([]*domain.Room, error) {
	if keyword != "" {
		return r.roomRepo.SearchByName(ctx, keyword)
	}
	if publicOnly {
		return r.roomRepo.ListPublic(ctx, limit, offset)
	}
	return r.roomRepo.ListAll(ctx, limit, offset)
}

// UpdateRoom implements RoomService.
func (r *roomService) UpdateRoom(ctx context.Context, actorID string, room *domain.Room) error {
	existingRoom, err := r.roomRepo.GetByID(ctx, room.ID)
	if err != nil {
		return fmt.Errorf("sala não encontrada: %w", err)
	}

	member, err := r.memberRepo.GetMemberByID(ctx, room.ID, actorID)
	if err != nil {
		return fmt.Errorf("usuario %s não é membro da sala", actorID)
	}
	if member.Role != domain.RoleOwner && member.Role != domain.RoleAdmin {
		return fmt.Errorf("usuario %s não tem permissão para atualizar a sala", actorID)
	}

	existingRoom.Name = room.Name
	existingRoom.Description = room.Description
	existingRoom.UpdatedAt = time.Now()

	return r.roomRepo.Update(ctx, existingRoom)
}

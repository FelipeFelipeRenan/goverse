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
	db         DBPool
	roomRepo   repository.RoomRepository
	memberRepo repository.RoomMemberRepository
}

func NewRoomService(db DBPool, r repository.RoomRepository, m repository.RoomMemberRepository) RoomService {
	return &roomService{
		db:         db,
		roomRepo:   r,
		memberRepo: m,
	}
}

func (s *roomService) CreateRoom(ctx context.Context, ownerID string, room *domain.Room) (*domain.Room, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	defer tx.Rollback(ctx)

	room.OwnerID = ownerID
	room.MemberCount = 0

	if err := s.roomRepo.Create(ctx, tx, room); err != nil {
		return nil, err
	}

	member := &domain.RoomMember{
		RoomID:   room.ID,
		UserID:   ownerID,
		Role:     domain.RoleOwner,
		JoinedAt: room.CreatedAt,
	}
	if err := s.memberRepo.AddMember(ctx, tx, member); err != nil {
		return nil, err
	}

	if err := s.roomRepo.IncrementMemberCount(ctx, tx, room.ID, 1); err != nil {
		return nil, err
	}
	room.MemberCount = 1

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("erro ao commitar transação: %w", err)
	}

	return room, nil
}

func (s *roomService) DeleteRoom(ctx context.Context, actorID string, roomID string) error {
	room, err := s.roomRepo.GetByID(ctx, s.db, roomID)
	if err != nil {
		return fmt.Errorf("sala não encontrada: %w", err)
	}
	if room.OwnerID != actorID {
		return domain.ErrUnauthorized
	}

	return s.roomRepo.Delete(ctx, s.db, roomID)
}

func (s *roomService) GetRoomByID(ctx context.Context, roomID string) (*domain.Room, error) {
	return s.roomRepo.GetByID(ctx, s.db, roomID)
}

func (s *roomService) ListRooms(ctx context.Context, limit int, offset int, publicOnly bool, keyword string) ([]*domain.Room, error) {
	if keyword != "" {
		return s.roomRepo.SearchByName(ctx, s.db, keyword)
	}
	if publicOnly {
		return s.roomRepo.ListPublic(ctx, s.db, limit, offset)
	}
	return s.roomRepo.ListAll(ctx, s.db, limit, offset)
}

func (s *roomService) UpdateRoom(ctx context.Context, actorID string, room *domain.Room) error {
	existingRoom, err := s.roomRepo.GetByID(ctx, s.db, room.ID)
	if err != nil {
		return fmt.Errorf("sala não encontrada: %w", err)
	}

	member, err := s.memberRepo.GetMemberByID(ctx, s.db, room.ID, actorID)
	if err != nil {
		return fmt.Errorf("usuario %s não é membro da sala", actorID)
	}
	if member.Role != domain.RoleOwner && member.Role != domain.RoleAdmin {
		return domain.ErrUnauthorized
	}

	existingRoom.Name = room.Name
	existingRoom.Description = room.Description
	existingRoom.UpdatedAt = time.Now()

	return s.roomRepo.Update(ctx, s.db, existingRoom)
}

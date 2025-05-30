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
	AddMember(ctx context.Context, actorID, roomID, userID string, role domain.Role) error
	RemoveMember(ctx context.Context, actorID, roomID, userID string) error
	UpdateMemberRole(ctx context.Context, actorID, roomID, userID string, newRole domain.Role) error
	GetRoomByID(ctx context.Context, roomID string) (*domain.Room, error)
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

// AddMember implements RoomService.
func (r *roomService) AddMember(ctx context.Context, actorID string, roomID string, userID string, role domain.Role) error {

	existingMember, err := r.memberRepo.IsMember(ctx, roomID, userID)
	if err != nil && existingMember {
		return fmt.Errorf("usuario %s já é membro da sala %s", userID, roomID)
	}
	member := &domain.RoomMember{
		RoomID:   roomID,
		UserID:   userID,
		Role:     role,
		JoinedAt: time.Now(),
	}

	return r.memberRepo.AddMember(ctx, member)
}

// UpdateMemberRole implements RoomService.
func (r *roomService) UpdateMemberRole(ctx context.Context, actorID string, roomID string, userID string, newRole domain.Role) error {

	// verifica se actor é membro da sala
	actor, err := r.memberRepo.GetMemberByID(ctx, roomID, actorID)
	if err != nil {
		return fmt.Errorf("usuario %s não é membro da sala", actorID)
	}

	// verifica se actor tem permissão
	if actor.Role != domain.RoleOwner && actor.Role != domain.RoleAdmin {
		return fmt.Errorf("usuario %s não tem permissão para alterar cargos", actorID)
	}

	// Evita que se mude o cargo do owner
	target, err := r.memberRepo.GetMemberByID(ctx, roomID, userID)
	if err != nil {
		return fmt.Errorf("usuario %s não é membro da sala", userID)
	}
	if target.Role == domain.RoleOwner {
		return fmt.Errorf("não é possivel alterar o cargo do dono da sala")
	}

	// atualiza a role do membro
	return r.memberRepo.UpdateMemberRole(ctx, roomID, userID, newRole)
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

// RemoveMember implements RoomService.
func (r *roomService) RemoveMember(ctx context.Context, actorID string, roomID string, userID string) error {

	actor, err := r.memberRepo.GetMemberByID(ctx, roomID, actorID)
	if err != nil {
		return fmt.Errorf("usuario %s não é membro da sala", actorID)
	}

	// busca o alvo da remoção
	target, err := r.memberRepo.GetMemberByID(ctx, roomID, userID)
	if err != nil {
		return fmt.Errorf("usuario não é membro da sala")
	}

	// verifica as permissões
	switch actor.Role {
	case domain.RoleOwner:
		if target.UserID == actor.UserID {
			return fmt.Errorf("O dono da sala não pode se remover")
		}
	case domain.RoleAdmin:
		if target.Role == domain.RoleAdmin || target.Role == domain.RoleOwner {
			return fmt.Errorf("admins não podem remover outros admins ou o dono")
		}
	default:
		return fmt.Errorf("usuario %s não tem permissão para remover membros", actorID)

	}

	return r.memberRepo.RemoveMember(ctx, roomID, userID)

}

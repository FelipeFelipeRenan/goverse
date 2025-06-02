package service

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
	GetRoomMembers(ctx context.Context, roomID string) ([]*domain.RoomMember, error)

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

	// verifica se a sala existe
	room, err := m.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return err
	}

	// verifica se quem está tentando adicionar é o owner
	if room.OwnerID != actorID {
		return domain.ErrUnauthorized
	}

	// verifica se o usuario existe, via conexao grpc ao user-service
	exists, err := m.userClient.ExistsUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if !exists {
		return domain.ErrUserNotFound
	}

	// verifica se o usuario ja está na sala
	_, err = m.memberRepo.GetMemberByID(ctx, roomID, userID)
	if err == nil {
		return domain.ErrMemberAlreadyExists
	}

	// cria o membro
	member := domain.RoomMember{
		RoomID: roomID,
		UserID: userID,
		Role:   role,
	}

	if err := m.memberRepo.AddMember(ctx, &member); err != nil {
		return err
	}

	if err := m.roomRepo.IncrementMemberCount(ctx, roomID, 1); err != nil {
		return err
	}
	return nil

}

// RemoveMember implements MemberService.
func (m *memberService) RemoveMember(ctx context.Context, actorID string, roomID string, userID string) error {
	// verifica a sala
	room, err := m.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return err
	}

	// somente o dono da sala pode remover
	if actorID != userID && actorID != room.OwnerID {
		return domain.ErrUnauthorized
	}

	// nao pode remover o dono
	if userID == room.OwnerID {
		return domain.ErrCannotRemoveOwner
	}

	// verifica se o membro existe
	_, err = m.memberRepo.GetMemberByID(ctx, roomID, userID)
	if err != nil {
		return domain.ErrMemberNotFound
	}

	// remove o membro
	if err := m.memberRepo.RemoveMember(ctx, roomID, userID); err != nil {
		return err
	}

	// atualiza o member_count
	if err := m.roomRepo.IncrementMemberCount(ctx, roomID, -1); err != nil {
		return err
	}

	return nil

}

// UpdateMemberRole implements MemberService.
func (m *memberService) UpdateMemberRole(ctx context.Context, actorID string, roomID string, userID string, newRole domain.Role) error {

	// verifica a sala
	room, err := m.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return err
	}

	// somente o dono da sala pode remover
	if actorID != userID && actorID != room.OwnerID {
		return domain.ErrUnauthorized
	}

	// nao pode remover o dono
	if userID == room.OwnerID {
		return domain.ErrCannotRemoveOwner
	}

	// verifica se o membro existe
	member, err := m.memberRepo.GetMemberByID(ctx, roomID, userID)
	if err != nil {
		return domain.ErrMemberNotFound
	}

	member.Role = newRole
	return m.memberRepo.UpdateMemberRole(ctx, roomID, member.UserID, newRole)
}

// GetRoomMembers implements MemberService.
func (m *memberService) GetRoomMembers(ctx context.Context, roomID string) ([]*domain.RoomMember, error) {
	return m.memberRepo.GetMembers(ctx, roomID)
}

// IsUserValid implements MemberService.
func (m *memberService) IsUserValid(ctx context.Context, userID string) (bool, error) {
	return m.userClient.ExistsUserByID(ctx, userID)
}

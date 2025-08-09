package service

import (
	"context"
	"fmt"

	"github.com/FelipeFelipeRenan/goverse/room-service/internal/client"
	"github.com/FelipeFelipeRenan/goverse/room-service/internal/domain"
	"github.com/FelipeFelipeRenan/goverse/room-service/internal/dtos"
	"github.com/FelipeFelipeRenan/goverse/room-service/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MemberService interface {
	AddMember(ctx context.Context, actorID, roomID, userID string, role domain.Role) error
	RemoveMember(ctx context.Context, actorID, roomID, userID string) error
	UpdateMemberRole(ctx context.Context, actorID, roomID, userID string, newRole domain.Role) error
	GetRoomMembers(ctx context.Context, roomID string) ([]*dtos.MemberWithUser, error)
	GetRoomsByUserID(ctx context.Context, userID string) ([]*domain.Room, error)
	GetRoomsByOwnerID(ctx context.Context, userID string) ([]*domain.Room, error)
	JoinRoom(ctx context.Context, roomID, userID, inviteToken string) error
}

type memberService struct {
	db         *pgxpool.Pool
	memberRepo repository.RoomMemberRepository
	roomRepo   repository.RoomRepository
	userClient client.UserServiceClient
}

func NewMemberService(db *pgxpool.Pool, m repository.RoomMemberRepository, r repository.RoomRepository, u client.UserServiceClient) MemberService {
	return &memberService{
		db:         db,
		memberRepo: m,
		roomRepo:   r,
		userClient: u,
	}
}

func (s *memberService) AddMember(ctx context.Context, actorID string, roomID string, userID string, role domain.Role) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	defer tx.Rollback(ctx)

	room, err := s.roomRepo.GetByID(ctx, tx, roomID)
	if err != nil {
		return domain.ErrRoomNotFound
	}
	if room.OwnerID != actorID {
		return domain.ErrUnauthorized
	}
	exists, err := s.userClient.ExistsUserByID(ctx, userID)
	if err != nil || !exists {
		return domain.ErrUserNotFound
	}
	isAlreadyMember, err := s.memberRepo.IsMember(ctx, tx, roomID, userID)
	if err != nil {
		return err
	}
	if isAlreadyMember {
		return domain.ErrMemberAlreadyExists
	}

	member := &domain.RoomMember{RoomID: roomID, UserID: userID, Role: role}
	if err := s.memberRepo.AddMember(ctx, tx, member); err != nil {
		return err
	}
	if err := s.roomRepo.IncrementMemberCount(ctx, tx, roomID, 1); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *memberService) RemoveMember(ctx context.Context, actorID string, roomID string, userID string) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	defer tx.Rollback(ctx)

	room, err := s.roomRepo.GetByID(ctx, tx, roomID)
	if err != nil {
		return domain.ErrRoomNotFound
	}
	if actorID != userID && actorID != room.OwnerID {
		return domain.ErrUnauthorized
	}
	if userID == room.OwnerID {
		return domain.ErrCannotRemoveOwner
	}
	if err := s.memberRepo.RemoveMember(ctx, tx, roomID, userID); err != nil {
		return err
	}
	if err := s.roomRepo.IncrementMemberCount(ctx, tx, roomID, -1); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *memberService) UpdateMemberRole(ctx context.Context, actorID, roomID, userID string, newRole domain.Role) error {
	room, err := s.roomRepo.GetByID(ctx, s.db, roomID)
	if err != nil {
		return domain.ErrRoomNotFound
	}
	if actorID != room.OwnerID {
		return domain.ErrUnauthorized
	}
	if userID == room.OwnerID {
		return domain.ErrCannotUpdateOwnerRole
	}

	return s.memberRepo.UpdateMemberRole(ctx, s.db, roomID, userID, newRole)
}

func (s *memberService) GetRoomMembers(ctx context.Context, roomID string) ([]*dtos.MemberWithUser, error) {
	members, err := s.memberRepo.GetMembers(ctx, s.db, roomID)
	if err != nil {
		return nil, err
	}

	var enrichedMembers []*dtos.MemberWithUser
	for _, member := range members {
		userResp, err := s.userClient.GetUserByID(ctx, member.UserID)
		if err != nil {
			continue
		}
		enrichedMembers = append(enrichedMembers, &dtos.MemberWithUser{
			RoomID:   member.RoomID,
			Role:     string(member.Role),
			JoinedAt: member.JoinedAt,
			User: dtos.UserDTO{
				ID:        userResp.Id,
				Name:      userResp.Name,
				Email:     userResp.Email,
				Picture:   userResp.Picture,
				CreatedAt: userResp.CreatedAt,
				IsOAuth:   userResp.IsOauth,
			},
		})
	}
	return enrichedMembers, nil
}

func (s *memberService) GetRoomsByUserID(ctx context.Context, userID string) ([]*domain.Room, error) {
	return s.memberRepo.GetRoomsByUserID(ctx, s.db, userID)
}

func (s *memberService) GetRoomsByOwnerID(ctx context.Context, userID string) ([]*domain.Room, error) {
	return s.memberRepo.GetRoomsByOwnerID(ctx, s.db, userID)
}

func (s *memberService) JoinRoom(ctx context.Context, roomID, userID, inviteToken string) error {
	// A lógica para entrar em uma sala também deveria ser transacional para ser mais robusta
	// Mas por enquanto, vamos mantê-la simples para não complicar
	isAlreadyMember, err := s.memberRepo.IsMember(ctx, s.db, roomID, userID)
	if err != nil {
		return err
	}
	if isAlreadyMember {
		return nil // Já é membro, sucesso silencioso
	}

	room, err := s.roomRepo.GetByID(ctx, s.db, roomID)
	if err != nil {
		return domain.ErrRoomNotFound
	}
	if !room.IsPublic {
		return domain.ErrForbidden
	}

	// Reutiliza a lógica transacional do AddMember, agindo como o dono da sala para adicionar um novo membro
	return s.AddMember(ctx, room.OwnerID, roomID, userID, domain.RoleMember)
}

package repository

import (
	"context"
	"errors"
	"time"

	"github.com/FelipeFelipeRenan/goverse/room-service/internal/domain"
	"github.com/jackc/pgx/v5"
)

//go:generate mockery --name RoomMemberRepository --output ./mocks --outpkg mocks --filename mock_member_room_repository.go
type RoomMemberRepository interface {
	AddMember(ctx context.Context, dbtx DBTX, member *domain.RoomMember) error
	RemoveMember(ctx context.Context, dbtx DBTX, roomID, userID string) error
	GetMembers(ctx context.Context, dbtx DBTX, roomID string) ([]*domain.RoomMember, error)
	GetMemberByID(ctx context.Context, dbtx DBTX, roomID, userID string) (*domain.RoomMember, error)
	UpdateMemberRole(ctx context.Context, dbtx DBTX, roomID, userID string, newRole domain.Role) error
	GetRoomsByUserID(ctx context.Context, dbtx DBTX, userID string) ([]*domain.Room, error)
	GetRoomsByOwnerID(ctx context.Context, dbtx DBTX, userID string) ([]*domain.Room, error)
	IsMember(ctx context.Context, dbtx DBTX, roomID, userID string) (bool, error)
}

type roomMemberRepository struct{}

func NewRoomMemberRepository() RoomMemberRepository {
	return &roomMemberRepository{}
}

func (r *roomMemberRepository) AddMember(ctx context.Context, dbtx DBTX, member *domain.RoomMember) error {
	query := `
		INSERT INTO room_members (room_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, $4)
	`
	if member.JoinedAt.IsZero() {
		member.JoinedAt = time.Now()
	}
	_, err := dbtx.Exec(ctx, query, member.RoomID, member.UserID, member.Role, member.JoinedAt)
	return err
}

func (r *roomMemberRepository) RemoveMember(ctx context.Context, dbtx DBTX, roomID string, userID string) error {
	query := `DELETE FROM room_members WHERE room_id = $1 AND user_id = $2`
	cmdTag, err := dbtx.Exec(ctx, query, roomID, userID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrMemberNotFound
	}
	return nil
}

func (r *roomMemberRepository) GetMemberByID(ctx context.Context, dbtx DBTX, roomID string, userID string) (*domain.RoomMember, error) {
	query := `SELECT room_id, user_id, role, joined_at FROM room_members WHERE room_id = $1 AND user_id = $2`
	var member domain.RoomMember
	err := dbtx.QueryRow(ctx, query, roomID, userID).Scan(&member.RoomID, &member.UserID, &member.Role, &member.JoinedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrMemberNotFound
		}
		return nil, err
	}
	return &member, nil
}

func (r *roomMemberRepository) GetMembers(ctx context.Context, dbtx DBTX, roomID string) ([]*domain.RoomMember, error) {
	query := `SELECT user_id, role, joined_at FROM room_members WHERE room_id = $1`
	rows, err := dbtx.Query(ctx, query, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*domain.RoomMember
	for rows.Next() {
		var m domain.RoomMember
		m.RoomID = roomID
		var roleStr string
		if err := rows.Scan(&m.UserID, &roleStr, &m.JoinedAt); err != nil {
			return nil, err
		}
		m.Role = domain.Role(roleStr)
		members = append(members, &m)
	}
	return members, nil
}

func (r *roomMemberRepository) UpdateMemberRole(ctx context.Context, dbtx DBTX, roomID string, userID string, newRole domain.Role) error {
	query := `UPDATE room_members SET role = $1 WHERE room_id = $2 AND user_id = $3`
	cmdTag, err := dbtx.Exec(ctx, query, newRole, roomID, userID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrMemberNotFound
	}
	return nil
}

func (r *roomMemberRepository) GetRoomsByUserID(ctx context.Context, dbtx DBTX, userID string) ([]*domain.Room, error) {
	query := `
		SELECT r.id, r.owner_id, r.name, r.description, r.member_count, r.max_members, r.is_public, r.created_at, r.updated_at
		FROM rooms r
		INNER JOIN room_members rm ON rm.room_id = r.id
		WHERE rm.user_id = $1 AND r.deleted_at IS NULL
	`
	rows, err := dbtx.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToAddrOfStructByPos[domain.Room])
}

func (r *roomMemberRepository) IsMember(ctx context.Context, dbtx DBTX, roomID string, userID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM room_members WHERE room_id = $1 AND user_id = $2)`
	var exists bool
	err := dbtx.QueryRow(ctx, query, roomID, userID).Scan(&exists)
	return exists, err
}

func (r *roomMemberRepository) GetRoomsByOwnerID(ctx context.Context, dbtx DBTX, userID string) ([]*domain.Room, error) {
	query := `
		SELECT id, owner_id, name, description, member_count, max_members, is_public, created_at, updated_at
		FROM rooms
		WHERE owner_id = $1 AND deleted_at IS NULL
	`
	rows, err := dbtx.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToAddrOfStructByPos[domain.Room])
}

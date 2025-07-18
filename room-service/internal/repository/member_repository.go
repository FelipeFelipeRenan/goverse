package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/FelipeFelipeRenan/goverse/room-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoomMemberRepository interface {
	AddMember(ctx context.Context, member *domain.RoomMember) error
	RemoveMember(ctx context.Context, roomID, userID string) error
	GetMembers(ctx context.Context, roomID string) ([]*domain.RoomMember, error)
	GetMemberByID(ctx context.Context, roomID, userID string) (*domain.RoomMember, error)
	GetMemberRole(ctx context.Context, roomID, userID string) (domain.Role, error)
	UpdateMemberRole(ctx context.Context, roomID, userID string, newRole domain.Role) error
	GetRoomsByUserID(ctx context.Context, userID string) ([]*domain.Room, error)
	GetRoomsByOwnerID(ctx context.Context, userID string) ([]*domain.Room, error)
	IsMember(ctx context.Context, roomID, userID string) (bool, error)
}

type roomMemberRepository struct {
	db *pgxpool.Pool
}

func NewRoomMemberRepository(db *pgxpool.Pool) RoomMemberRepository {
	return &roomMemberRepository{db: db}
}

// AddMember implements RoomMemberRepository.
func (r *roomMemberRepository) AddMember(ctx context.Context, member *domain.RoomMember) error {
	query := `
		INSERT INTO room_members (room_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (room_id, user_id) DO UPDATE SET role = $3
	`

	if member.JoinedAt.IsZero() {
		member.JoinedAt = time.Now()
	}

	_, err := r.db.Exec(context.Background(), query,
		member.RoomID,
		member.UserID,
		member.Role,
		member.JoinedAt,
	)
	return err
}

// RemoveMember implements RoomMemberRepository.
func (r *roomMemberRepository) RemoveMember(ctx context.Context, roomID string, userID string) error {
	query := `DELETE FROM room_members WHERE room_id = $1 AND user_id = $2`
	_, err := r.db.Exec(ctx, query, roomID, userID)
	return err
}

// GetMemberByID implements RoomMemberRepository.
func (r *roomMemberRepository) GetMemberByID(ctx context.Context, roomID string, userID string) (*domain.RoomMember, error) {
	query := `
		SELECT room_id, user_id, role, joined_at
		FROM room_members
		WHERE room_id = $1 AND user_id = $2
	`

	var member domain.RoomMember
	err := r.db.QueryRow(ctx, query, roomID, userID).Scan(
		&member.RoomID,
		&member.UserID,
		&member.Role,
		&member.JoinedAt,
	)
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// GetMembers implements RoomMemberRepository.
func (r *roomMemberRepository) GetMembers(ctx context.Context, roomID string) ([]*domain.RoomMember, error) {
	query := `
		SELECT user_id, role, joined_at
		FROM room_members
		WHERE room_id = $1
	`

	rows, err := r.db.Query(ctx, query, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*domain.RoomMember
	for rows.Next() {
		var m domain.RoomMember
		m.RoomID = roomID
		var roleStr string
		err := rows.Scan(&m.UserID, &roleStr, &m.JoinedAt)
		if err != nil {
			return nil, err
		}
		m.Role = domain.Role(roleStr)
		members = append(members, &m)
	}

	return members, nil
}

// GetUserRole implements RoomMemberRepository.
func (r *roomMemberRepository) GetMemberRole(ctx context.Context, roomID string, userID string) (domain.Role, error) {
	query := `SELECT role FROM room_members WHERE room_id = $1 AND user_id = $2`

	var roleStr string
	err := r.db.QueryRow(ctx, query, roomID, userID).Scan(roleStr)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return domain.Role(roleStr), nil
}

// UpdateMemberRole implements RoomMemberRepository.
func (r *roomMemberRepository) UpdateMemberRole(ctx context.Context, roomID string, userID string, newRole domain.Role) error {
	query := `
		UPDATE room_members
		SET role = $1
		WHERE room_id = $2 AND user_id = $3
	`

	cmdTag, err := r.db.Exec(ctx, query, newRole, roomID, userID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("nenhum membro encontrado com room_id=%s e user_id=%s", userID, roomID)
	}
	return nil
}

func (r *roomMemberRepository) GetRoomsByUserID(ctx context.Context, userID string) ([]*domain.Room, error) {
	rows, err := r.db.Query(ctx, `
		SELECT r.id, r.owner_id, r.name, r.description, r.member_count, r.max_members, r.created_at, r.updated_at
		FROM rooms r
		INNER JOIN room_members rm ON rm.room_id = r.id
		WHERE rm.user_id = $1  AND r.deleted_at IS NULL
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []*domain.Room
	for rows.Next() {
		var room domain.Room
		if err := rows.Scan(&room.ID, &room.OwnerID, &room.Name, &room.Description, &room.MemberCount, &room.MaxMembers, &room.CreatedAt, &room.UpdatedAt); err != nil {
			return nil, err
		}
		rooms = append(rooms, &room)
	}

	return rooms, nil
}

// IsMember implements RoomMemberRepository.
func (r *roomMemberRepository) IsMember(ctx context.Context, roomID string, userID string) (bool, error) {
	query := `SELECT 1 FROM room_members WHERE room_id = $1 AND user_id = $2`

	var dummy int
	err := r.db.QueryRow(context.Background(), query, roomID, userID).Scan(&dummy)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *roomMemberRepository) GetRoomsByOwnerID(ctx context.Context, userID string) ([]*domain.Room, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, owner_id, name, description, member_count, max_members, created_at, updated_at
		FROM rooms
		WHERE owner_id = $1 AND deleted_at IS NULL
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []*domain.Room
	for rows.Next() {
		var room domain.Room
		if err := rows.Scan(&room.ID, &room.OwnerID, &room.Name, &room.Description, &room.MemberCount, &room.MaxMembers, &room.CreatedAt, &room.UpdatedAt); err != nil {
			return nil, err
		}
		rooms = append(rooms, &room)
	}

	return rooms, nil
}

package repository

import (
	"context"
	"errors"
	"time"

	"github.com/FelipeFelipeRenan/goverse/room-service/internal/domain"
	"github.com/jackc/pgx/v5"
)

type RoomMemberRepository interface {
	AddMember(ctx context.Context, member *domain.RoomMember) error
	RemoveMember(roomID, userID string) error
	GetMembers(roomID string) ([]*domain.RoomMember, error)
	GetUserRole(roomID, userID string) (domain.Role, error)
	IsMember(roomID, userID string) (bool, error)
}

type roomMemberRepository struct {
	db *pgx.Conn
}

func NewRoomMemberRepository(db *pgx.Conn) RoomMemberRepository {
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
func (r *roomMemberRepository) RemoveMember(roomID string, userID string) error {
	query := `DELETE FROM room_members WHERE room_id = $1 AND user_id = $2`
	_, err := r.db.Exec(context.Background(), query, roomID, userID)
	return err
}

// GetMembers implements RoomMemberRepository.
func (r *roomMemberRepository) GetMembers(roomID string) ([]*domain.RoomMember, error) {
	query := `
		SELECT user_id, role, joined_at
		FROM room_members
		WHERE room_id = $1
	`

	rows, err := r.db.Query(context.Background(), query, roomID)
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
func (r *roomMemberRepository) GetUserRole(roomID string, userID string) (domain.Role, error) {
	query := `SELECT role FROM room_members WHERE room_id = $1 AND user_id = $2`

	var roleStr string
	err := r.db.QueryRow(context.Background(), query, roomID, userID).Scan(roleStr)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return domain.Role(roleStr), nil
}

// IsMember implements RoomMemberRepository.
func (r *roomMemberRepository) IsMember(roomID string, userID string) (bool, error) {
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

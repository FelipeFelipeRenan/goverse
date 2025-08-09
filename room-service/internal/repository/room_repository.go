package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/FelipeFelipeRenan/goverse/room-service/internal/domain"
	"github.com/jackc/pgx/v5"
)

//go:generate mockery --name RoomRepository --output ./mocks --outpkg mocks --filename mock_room_repository.go
type RoomRepository interface {
	Create(ctx context.Context, dbtx DBTX, room *domain.Room) error
	GetByID(ctx context.Context, dbtx DBTX, id string) (*domain.Room, error)
	ListAll(ctx context.Context, dbtx DBTX, limit, offset int) ([]*domain.Room, error)
	ListPublic(ctx context.Context, dbtx DBTX, limit, offset int) ([]*domain.Room, error)
	Update(ctx context.Context, dbtx DBTX, room *domain.Room) error
	Delete(ctx context.Context, dbtx DBTX, id string) error
	SearchByName(ctx context.Context, dbtx DBTX, keyword string) ([]*domain.Room, error)
	IncrementMemberCount(ctx context.Context, dbtx DBTX, roomID string, delta int) error
}

type roomRepository struct{}

func NewRoomRepository() RoomRepository {
	return &roomRepository{}
}

func (r *roomRepository) Create(ctx context.Context, dbtx DBTX, room *domain.Room) error {
	query := `
		INSERT INTO rooms (name, description, is_public, owner_id, member_count, max_members, created_at, updated_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at
	`
	return dbtx.QueryRow(ctx, query,
		room.Name, room.Description, room.IsPublic, room.OwnerID,
		room.MemberCount, room.MaxMembers, time.Now(), time.Now(), nil,
	).Scan(&room.ID, &room.CreatedAt)
}

func (r *roomRepository) Delete(ctx context.Context, dbtx DBTX, id string) error {
	query := `UPDATE rooms SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	cmdTag, err := dbtx.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrRoomNotFound
	}
	return nil
}

func (r *roomRepository) Update(ctx context.Context, dbtx DBTX, room *domain.Room) error {
	query := `
		UPDATE rooms
		SET name = $1, description = $2, is_public = $3, max_members = $4, updated_at = $5
		WHERE id = $6 AND deleted_at IS NULL
	`
	cmdTag, err := dbtx.Exec(ctx, query,
		room.Name, room.Description, room.IsPublic, room.MaxMembers, time.Now(), room.ID,
	)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrRoomNotFound
	}
	return nil
}

func (r *roomRepository) GetByID(ctx context.Context, dbtx DBTX, id string) (*domain.Room, error) {
	query := `
		SELECT id, name, description, is_public, owner_id, member_count, max_members, created_at, updated_at
		FROM rooms
		WHERE id = $1 AND deleted_at IS NULL
	`
	row := dbtx.QueryRow(ctx, query, id)
	var room domain.Room
	err := row.Scan(
		&room.ID, &room.Name, &room.Description, &room.IsPublic, &room.OwnerID,
		&room.MemberCount, &room.MaxMembers, &room.CreatedAt, &room.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrRoomNotFound
		}
		return nil, err
	}
	return &room, nil
}

func (r *roomRepository) ListAll(ctx context.Context, dbtx DBTX, limit, offset int) ([]*domain.Room, error) {
	query := `
        SELECT id, name, description, owner_id, is_public, member_count, max_members, created_at, updated_at
        FROM rooms WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2
    `
	rows, err := dbtx.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToAddrOfStructByPos[domain.Room])
}

func (r *roomRepository) ListPublic(ctx context.Context, dbtx DBTX, limit, offset int) ([]*domain.Room, error) {
	query := `
		SELECT id, name, description, owner_id, is_public, member_count, max_members, created_at, updated_at
		FROM rooms WHERE is_public = true AND deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`
	rows, err := dbtx.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToAddrOfStructByPos[domain.Room])
}

func (r *roomRepository) SearchByName(ctx context.Context, dbtx DBTX, keyword string) ([]*domain.Room, error) {
	query := `
		SELECT id, name, description, is_public, owner_id, member_count, max_members, created_at, updated_at
		FROM rooms WHERE name ILIKE '%' || $1 || '%' AND deleted_at IS NULL
	`
	rows, err := dbtx.Query(ctx, query, keyword)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToAddrOfStructByPos[domain.Room])
}

func (r *roomRepository) IncrementMemberCount(ctx context.Context, dbtx DBTX, roomID string, delta int) error {
	query := `UPDATE rooms SET member_count = member_count + $1, updated_at = $2 WHERE id = $3 AND deleted_at IS NULL`
	cmdTag, err := dbtx.Exec(ctx, query, delta, time.Now(), roomID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("tentativa de incrementar membro de sala não encontrada: %s", roomID)
	}
	return nil
}

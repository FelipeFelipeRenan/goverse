package repository

import (
	"context"
	"errors"
	"time"

	"github.com/FelipeFelipeRenan/goverse/room-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoomRepository interface {
	Create(ctx context.Context, room *domain.Room) error
	GetByID(ctx context.Context, id string) (*domain.Room, error)
	ListPublic(ctx context.Context) ([]*domain.Room, error)
	ListByUserID(ctx context.Context, userID string) ([]*domain.Room, error)
	Update(ctx context.Context, room *domain.Room) error
	Delete(ctx context.Context, id string) error
	Exists(ctx context.Context, id string) (bool, error)
	SearchByName(ctx context.Context, keyword string) ([]*domain.Room, error)
	IncrementMemberCount(ctx context.Context, roomID string, delta int) error
}

type roomRepository struct {
	db *pgxpool.Pool
}

func NewRoomRepository(db *pgxpool.Pool) RoomRepository {
	return &roomRepository{db: db}
}

// Create implements RoomRepository.

func (r *roomRepository) Create(ctx context.Context, room *domain.Room) error {
	query := `
		INSERT INTO rooms (name, description, is_public, owner_id, member_count, max_members,
		  created_at, updated_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	now := time.Now()
	room.CreatedAt = now
	room.UpdatedAt = now

	err := r.db.QueryRow(ctx, query,
		room.Name,
		room.Description,
		room.IsPublic,
		room.OwnerID,
		room.MemberCount,
		room.MaxMembers,
		room.CreatedAt,
		room.UpdatedAt,
		room.DeletedAt,
	).Scan(&room.ID)

	return err
}

// Delete implements RoomRepository.
func (r *roomRepository) Delete(ctx context.Context, id string) error {
	query := ` UPDATE rooms
				SET deleted_at = $1, updated_at = $1
				WHERE id = $2 AND deleted_at IS NULL `

	_, err := r.db.Exec(ctx, query, time.Now(), id)
	return err
}

// Update implements RoomRepository.
func (r *roomRepository) Update(ctx context.Context, room *domain.Room) error {
	room.UpdatedAt = time.Now()

	query := `
		UPDATE rooms
		SET name = $1, description = $2, is_public = $3, member_count = $4, max_members = $5, updated_at = $6
		WHERE id = $5 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, query,
		room.Name,
		room.Description,
		room.IsPublic,
		room.MemberCount,
		room.MaxMembers,
		room.UpdatedAt,
		room.ID,
	)
	return err
}

// GetByID implements RoomRepository.
func (r *roomRepository) GetByID(ctx context.Context, id string) (*domain.Room, error) {
	query := `
		SELECT id, name, description, is_public, owner_id, member_count, max_members, 
		 created_at, updated_at, deleted_at
		FROM rooms
		WHERE id = $1 AND deleted_at IS NULL
	`
	row := r.db.QueryRow(ctx, query, id)

	var room domain.Room

	err := row.Scan(
		&room.ID,
		&room.Name,
		&room.Description,
		&room.IsPublic,
		&room.OwnerID,
		&room.MemberCount,
		&room.MaxMembers,
		&room.CreatedAt,
		&room.UpdatedAt,
		&room.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &room, nil

}

// ListByUserID implements RoomRepository.
func (r *roomRepository) ListByUserID(ctx context.Context, userID string) ([]*domain.Room, error) {
	query := `
		SELECT r.id, r.name, r.description, r.is_public, r.owner_id, r.member_count, r.max_members ,
			r.created_at, r.updated_at, r.deleted_at
		FROM rooms r
		JOIN room_members m ON r.id = m.room_id
		WHERE m.user_id = $1 AND r.deleted_at IS NULL
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	var rooms []*domain.Room

	for rows.Next() {
		var room domain.Room
		err := rows.Scan(
			&room.ID,
			&room.Name,
			&room.Description,
			&room.IsPublic,
			&room.OwnerID,
			&room.MemberCount,
			&room.MaxMembers,
			&room.CreatedAt,
			&room.UpdatedAt,
			&room.DeletedAt,
		)

		if err != nil {
			return nil, err
		}
		rooms = append(rooms, &room)
	}
	return rooms, nil
}

// ListPublic implements RoomRepository.
func (r *roomRepository) ListPublic(ctx context.Context) ([]*domain.Room, error) {
	query := `
		SELECT id, name, description, is_public, owner_id, member_count, max_members,
		 created_at, updated_at, deleted_at
		FROM rooms
		WHERE is_public = true AND deleted_at IS NULL
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []*domain.Room

	for rows.Next() {
		var room domain.Room
		err := rows.Scan(
			&room.ID,
			&room.Name,
			&room.Description,
			&room.IsPublic,
			&room.OwnerID,
			&room.MemberCount,
			&room.MaxMembers,
			&room.CreatedAt,
			&room.UpdatedAt,
			&room.DeletedAt,
		)

		if err != nil {
			return nil, err
		}
		rooms = append(rooms, &room)
	}
	return rooms, nil

}

func (r *roomRepository) Exists(ctx context.Context, id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM rooms WHERE id = $1 AND deleted_at IS NULL)`
	var exists bool
	err := r.db.QueryRow(ctx, query, id).Scan(&exists)
	return exists, err
}

// SearchByName implements RoomRepository.
func (r *roomRepository) SearchByName(ctx context.Context, keyword string) ([]*domain.Room, error) {
	query := `
		SELECT id, name, description, is_public, owner_id,
		       member_count, max_members, created_at, updated_at, deleted_at
		FROM rooms
		WHERE name ILIKE '%' || $1 || '%' AND deleted_at IS NULL
	`

	rows, err := r.db.Query(ctx, query, keyword)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []*domain.Room

	for rows.Next() {
		var room domain.Room
		err := rows.Scan(
			&room.ID, &room.Name, &room.Description, &room.IsPublic,
			&room.OwnerID, &room.MemberCount, &room.MaxMembers,
			&room.CreatedAt, &room.UpdatedAt, &room.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, &room)
	}
	return rooms, nil
}

// IncrementMemberCount implements RoomRepository.
func (r *roomRepository) IncrementMemberCount(ctx context.Context, roomID string, delta int) error {
	panic("unimplemented")
}

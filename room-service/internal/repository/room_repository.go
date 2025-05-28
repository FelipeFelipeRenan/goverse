package repository

import (
	"context"
	"errors"
	"time"

	"github.com/FelipeFelipeRenan/goverse/room-service/internal/domain"
	"github.com/jackc/pgx/v5"
)

type RoomRepository interface {
	Create(room *domain.Room) error
	GetByID(id string) (*domain.Room, error)
	ListPublic() ([]*domain.Room, error)
	ListByUserID(userID string) ([]*domain.Room, error)
	Update(room *domain.Room) error
	Delete(id string) error
}

type roomRepository struct {
	db *pgx.Conn
}

func NewRoomRepository(db *pgx.Conn) RoomRepository {
	return &roomRepository{db: db}
}

// Create implements RoomRepository.
func (r *roomRepository) Create(room *domain.Room) error {
	query := `
		INSERT INTO rooms (id, name, description, is_public, owner_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	now := time.Now()

	room.CreatedAt = now
	room.UpdatedAt = now

	_, err := r.db.Exec(context.Background(), query,
		room.ID,
		room.Name,
		room, room.Description,
		room.IsPublic,
		room.OwnerID,
		room.CreatedAt,
		room.UpdatedAt,
	)
	return err

}

// Delete implements RoomRepository.
func (r *roomRepository) Delete(id string) error {
	query := `DELETE FROM rooms WHERE id = $1`

	_, err := r.db.Exec(context.Background(), query, id)
	return err
}

// Update implements RoomRepository.
func (r *roomRepository) Update(room *domain.Room) error {
	room.UpdatedAt = time.Now()

	query := `
		UPDATE rooms
		SET name = $1, description = $2, is_public = $3, updated_at = $4
		WHERE id = $5
	`

	_, err := r.db.Exec(context.Background(), query,
		room.Name,
		room.Description,
		room.IsPublic,
		room.UpdatedAt,
		room.ID,
	)
	return err
}

// GetByID implements RoomRepository.
func (r *roomRepository) GetByID(id string) (*domain.Room, error) {
	query := `
		SELECT id, name, description, is_public, owner_id, created_at, updated_at
		FROM rooms
		WHERE id = $1
	`
	row := r.db.QueryRow(context.Background(), query, id)

	var room domain.Room

	err := row.Scan(
		&room.ID,
		&room.Name,
		&room.Description,
		&room.IsPublic,
		&room.OwnerID,
		&room.CreatedAt,
		&room.UpdatedAt,
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
func (r *roomRepository) ListByUserID(userID string) ([]*domain.Room, error) {
	query := `
		SELECT r.id, r.name, r.description, r.is_public, r.owner_id, r.created_at, r.updated_at
		FROM rooms r
		JOIN room_members m ON r.id = m.room_id
		WHERE m.user_id = $1
	`

	rows, err := r.db.Query(context.Background(), query, userID)
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
			&room.CreatedAt,
			&room.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}
		rooms = append(rooms, &room)
	}
	return rooms, nil
}

// ListPublic implements RoomRepository.
func (r *roomRepository) ListPublic() ([]*domain.Room, error) {
	query := `
		SELECT id, name, description, is_public, owner_id, created_at, updated_at
		FROM rooms
		WHERE is_public = true
	`
	rows, err := r.db.Query(context.Background(), query)
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
			&room.CreatedAt,
			&room.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}
		rooms = append(rooms, &room)
	}
	return rooms, nil

}

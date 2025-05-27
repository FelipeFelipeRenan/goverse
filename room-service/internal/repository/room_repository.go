package repository

import (
	"context"
	"time"

	"github.com/FelipeFelipeRenan/goverse/room-service/internal/domain"
	"github.com/jackc/pgx/v5"
)

type RoomRepository interface {
	Create(room *domain.Room) error
	GetByID(id string) *domain.Room
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
	panic("unimplemented")
}

// GetByID implements RoomRepository.
func (r *roomRepository) GetByID(id string) *domain.Room {
	panic("unimplemented")
}

// ListByUserID implements RoomRepository.
func (r *roomRepository) ListByUserID(userID string) ([]*domain.Room, error) {
	panic("unimplemented")
}

// ListPublic implements RoomRepository.
func (r *roomRepository) ListPublic() ([]*domain.Room, error) {
	panic("unimplemented")
}

// Update implements RoomRepository.
func (r *roomRepository) Update(room *domain.Room) error {
	panic("unimplemented")
}

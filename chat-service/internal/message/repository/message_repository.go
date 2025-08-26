package repository

import (
	"context"

	"github.com/FelipeFelipeRenan/goverse/chat-service/internal/message/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MessageRepository interface {
	SaveMessage(ctx context.Context, msg *domain.Message) error
}

type pgxMessageRepository struct {
	db *pgxpool.Pool
}

func NewMessageRepository(db *pgxpool.Pool) MessageRepository {
	return &pgxMessageRepository{db: db}
}

// SaveMessage implements MessageRepository.
func (p *pgxMessageRepository) SaveMessage(ctx context.Context, msg *domain.Message) error {

	query := `        INSERT INTO messages (room_id, user_id, username, content)
        VALUES ($1, $2, $3, $4)`
	_, err := p.db.Exec(ctx, query, msg.RoomID, msg.UserID, msg.Username, msg.Content)

	return err
}

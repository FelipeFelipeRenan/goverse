package repository

import (
	"context"
	"fmt"
	"strconv"

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
func (r *pgxMessageRepository) SaveMessage(ctx context.Context, msg *domain.Message) error {
	// Converte RoomID de string para integer
	roomID, err := strconv.Atoi(msg.RoomID)
	if err != nil {
		return fmt.Errorf("roomID inválido, esperado um número: %s", msg.RoomID)
	}

	// Converte UserID de string para integer
	userID, err := strconv.Atoi(msg.UserID)
	if err != nil {
		return fmt.Errorf("userID inválido, esperado um número: %s", msg.UserID)
	}

	query := `
        INSERT INTO messages (room_id, user_id, username, content)
        VALUES ($1, $2, $3, $4)
    `
	// Passa os IDs convertidos para o banco
	_, err = r.db.Exec(ctx, query, roomID, userID, msg.Username, msg.Content)
	return err
}

package repository

import (
	"context"
	"fmt"

	"github.com/FelipeFelipeRenan/goverse/chat-service/internal/message/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MessageRepository interface {
	SaveMessage(ctx context.Context, msg *domain.Message) error
	GetMessagesByRoomID(ctx context.Context, roomID string, limit, offset int) ([]domain.Message, error)
}

type pgxMessageRepository struct {
	db *pgxpool.Pool
}

func NewMessageRepository(db *pgxpool.Pool) MessageRepository {
	return &pgxMessageRepository{db: db}
}

// SaveMessage implements MessageRepository.
func (r *pgxMessageRepository) SaveMessage(ctx context.Context, msg *domain.Message) error {

	query := `
        INSERT INTO messages (room_id, user_id, username, content)
        VALUES ($1, $2, $3, $4)
    `
	_, err := r.db.Exec(ctx, query, msg.RoomID, msg.UserID, msg.Username, msg.Content)
	if err != nil {
		return fmt.Errorf("falha ao salvar mensagem no banco: %w", err)
	}
	return nil
}

// GetMessagesByRoomID implements MessageRepository.
func (r *pgxMessageRepository) GetMessagesByRoomID(ctx context.Context, roomID string, limit int, offset int) ([]domain.Message, error) {
	query := `
        SELECT id, room_id, user_id, username, content, created_at 
        FROM messages 
        WHERE room_id = $1 
        ORDER BY created_at DESC 
        LIMIT $2 OFFSET $3
    `
	rows, err := r.db.Query(ctx, query, roomID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar mensagens no banco: %w", err)
	}
	defer rows.Close()

	var messages []domain.Message

	for rows.Next() {
		var msg domain.Message
		if err := rows.Scan(&msg.ID, &msg.RoomID, &msg.UserID, &msg.Username, &msg.Content, &msg.CreatedAt); err != nil {
			return nil, fmt.Errorf("erro ai escanear mensagem: %w", err)
		}

		messages = append(messages, msg)

		if rows.Err() != nil {
			return nil, rows.Err()
		}
	}
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, nil
}

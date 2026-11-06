package chat

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type store interface {
	// save atribui ID (UUIDv7) e CreatedAt gerados no servidor.
	save(ctx context.Context, m *Message) error
	// history devolve as mensagens CHAT mais recentes, em ordem cronológica.
	history(ctx context.Context, roomID string, limit, offset int) ([]Message, error)
}

type pgStore struct{ pool *pgxpool.Pool }

func (s *pgStore) save(ctx context.Context, m *Message) error {
	id, err := uuid.NewV7()
	if err != nil {
		return err
	}
	m.ID = id.String()
	m.CreatedAt = time.Now().UTC()
	_, err = s.pool.Exec(ctx,
		`INSERT INTO chat.messages (id, room_id, user_id, username, content, created_at) VALUES ($1,$2,$3,$4,$5,$6)`,
		id, m.RoomID, m.UserID, m.Username, m.Content, m.CreatedAt)
	if err != nil {
		return fmt.Errorf("salvar mensagem: %w", err)
	}
	return nil
}

func (s *pgStore) history(ctx context.Context, roomID string, limit, offset int) ([]Message, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, room_id, user_id, username, content, created_at
		   FROM chat.messages WHERE room_id = $1 ORDER BY id DESC LIMIT $2 OFFSET $3`,
		roomID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("buscar histórico: %w", err)
	}
	defer rows.Close()
	out := []Message{}
	for rows.Next() {
		m := Message{Type: TypeChat}
		var id, rid, uid uuid.UUID
		if err := rows.Scan(&id, &rid, &uid, &m.Username, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.ID, m.RoomID, m.UserID = id.String(), rid.String(), uid.String()
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 { // DESC → cronológica
		out[i], out[j] = out[j], out[i]
	}
	return out, nil
}

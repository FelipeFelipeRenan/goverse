package chat

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type store interface {
	// save grava a mensagem (ID e CreatedAt já definidos pelo chamador). Idempotente: reenvio do
	// mesmo ID é ignorado, o que torna seguro reprocessar o spool depois de um crash.
	save(ctx context.Context, m Message) error
	// deadLetter guarda uma mensagem que nunca poderá ser gravada.
	deadLetter(ctx context.Context, m Message, reason string) error
	byID(ctx context.Context, id string) (Message, error)
	// history devolve as mensagens CHAT mais recentes, em ordem cronológica.
	history(ctx context.Context, roomID string, limit, offset int) ([]Message, error)
}

type pgStore struct{ pool *pgxpool.Pool }

func newMessageID() (string, error) {
	id, err := uuid.NewV7()
	return id.String(), err
}

func (s *pgStore) save(ctx context.Context, m Message) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO chat.messages (id, room_id, user_id, username, content, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6) ON CONFLICT (id) DO NOTHING`,
		m.ID, m.RoomID, m.UserID, m.Username, m.Content, m.CreatedAt)
	if err != nil {
		return fmt.Errorf("salvar mensagem: %w", err)
	}
	return nil
}

func (s *pgStore) deadLetter(ctx context.Context, m Message, reason string) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO chat.dead_letters (id, room_id, user_id, username, content, sent_at, reason)
		 VALUES ($1,$2,$3,$4,$5,$6,$7) ON CONFLICT (id) DO NOTHING`,
		m.ID, m.RoomID, m.UserID, m.Username, m.Content, m.CreatedAt, reason)
	return err
}

func (s *pgStore) byID(ctx context.Context, id string) (Message, error) {
	m := Message{Type: TypeChat}
	var mid, rid, uid uuid.UUID
	err := s.pool.QueryRow(ctx,
		`SELECT id, room_id, user_id, username, content, created_at FROM chat.messages WHERE id = $1`, id).
		Scan(&mid, &rid, &uid, &m.Username, &m.Content, &m.CreatedAt)
	if err != nil {
		return Message{}, err
	}
	m.ID, m.RoomID, m.UserID = mid.String(), rid.String(), uid.String()
	return m, nil
}

// isPermanent: erro que repetir não resolve (violação de integridade, dado inválido).
// Qualquer outro (conexão, timeout, banco em recuperação) é transitório.
func isPermanent(err error) bool {
	var pe *pgconn.PgError
	if errors.As(err, &pe) {
		return strings.HasPrefix(pe.Code, "23") || strings.HasPrefix(pe.Code, "22")
	}
	return false
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

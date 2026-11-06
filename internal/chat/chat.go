// Package chat entrega mensagens em tempo real (WebSocket) e o histórico. Substitui o chat-service.
//
// Sem Kafka/Redis: um único processo mantém as conexões em memória. A mensagem é gravada no
// Postgres ANTES do fan-out, então o que o cliente vê já está persistido, com ordem definida.
// Para escalar para várias réplicas, o ponto de troca é o Hub (ver docs/chat-migration.md).
package chat

import (
	"context"
	"embed"
	"io/fs"
	"net/http"
	"time"

	"github.com/FelipeFelipeRenan/goverse/internal/platform/authn"
	"github.com/FelipeFelipeRenan/goverse/internal/platform/migrate"
	"github.com/FelipeFelipeRenan/goverse/internal/rooms"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Message mantém o formato JSON do chat-service antigo (inclusive "CreatedAt" sem snake_case),
// porque o frontend depende dele.
type Message struct {
	ID           string    `json:"id,omitempty"`
	Type         string    `json:"type,omitempty"`
	Content      string    `json:"content"`
	RoomID       string    `json:"room_id"`
	UserID       string    `json:"user_id"`
	Username     string    `json:"username"`
	TargetUserID string    `json:"target_user_id,omitempty"`
	CreatedAt    time.Time `json:",omitempty"`
}

const (
	TypeChat        = "CHAT"
	TypePresence    = "PRESENCE"
	TypeTypingStart = "TYPING_START"
	TypeTypingStop  = "TYPING_STOP"
	TypeSignal      = "SIGNAL"
	TypeDirect      = "DIRECT"

	maxContentBytes = 4000
)

type Config struct {
	// AllowedOrigins: origens de browser aceitas no handshake do WebSocket.
	AllowedOrigins []string
}

type Module struct {
	pool *pgxpool.Pool
	api  *api
}

func New(pool *pgxpool.Pool, keys *authn.Keys, members rooms.Membership, cfg Config) *Module {
	return newModule(pool, &pgStore{pool: pool}, keys, members, cfg)
}

func newModule(pool *pgxpool.Pool, st store, keys *authn.Keys, members rooms.Membership, cfg Config) *Module {
	h := newHub(st)
	return &Module{pool: pool, api: &api{hub: h, st: st, keys: keys, members: members, origins: toSet(cfg.AllowedOrigins)}}
}

func (m *Module) Name() string { return "chat" }

func (m *Module) Register(mux *http.ServeMux) { m.api.routes(mux) }

func (m *Module) Migrate(ctx context.Context) error {
	sub, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		return err
	}
	return migrate.Run(ctx, m.pool, "chat", sub)
}

func toSet(l []string) map[string]struct{} {
	s := make(map[string]struct{}, len(l))
	for _, v := range l {
		s[v] = struct{}{}
	}
	return s
}

// Package chat entrega mensagens em tempo real (WebSocket) e o histórico. Substitui o chat-service.
//
// Sem Kafka/Redis: um único processo mantém as conexões em memória. A mensagem é gravada no
// Postgres ANTES do fan-out, então o que o cliente vê já está persistido, com ordem definida.
// Para escalar para várias réplicas, o ponto de troca é o Hub (ver docs/chat-migration.md).
package chat

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"github.com/FelipeFelipeRenan/goverse/internal/platform/authn"
	"github.com/FelipeFelipeRenan/goverse/internal/platform/migrate"
	"github.com/FelipeFelipeRenan/goverse/internal/rooms"
	"github.com/google/uuid"
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
	// SpoolDir: se definido, mensagens que não puderam ser gravadas (banco fora) vão para um arquivo
	// local e são reprocessadas depois, em vez de serem recusadas. Precisa de disco que sobreviva a restart.
	SpoolDir string
	// Bus: "postgres" (padrão; LISTEN/NOTIFY, permite várias réplicas) ou "none" (processo único).
	Bus string
}

type Module struct {
	pool *pgxpool.Pool
	api  *api
}

func New(pool *pgxpool.Pool, keys *authn.Keys, members rooms.Membership, cfg Config) (*Module, error) {
	var opts hubOpts
	opts.origin = uuid.NewString()
	if cfg.SpoolDir != "" {
		sp, err := newSpool(cfg.SpoolDir)
		if err != nil {
			return nil, fmt.Errorf("spool: %w", err)
		}
		opts.spool = sp
	}
	if cfg.Bus != "none" {
		opts.bus = &pgBus{pool: pool}
	}
	st := &pgStore{pool: pool}
	return newModule(pool, st, keys, members, cfg.AllowedOrigins, newHub(st, opts)), nil
}

func newModule(pool *pgxpool.Pool, st store, keys *authn.Keys, members rooms.Membership, origins []string, h *hub) *Module {
	return &Module{pool: pool, api: &api{hub: h, st: st, keys: keys, members: members, origins: toSet(origins)}}
}

func (m *Module) Name() string { return "chat" }

func (m *Module) Register(mux *http.ServeMux) { m.api.routes(mux) }

// Run executa o trabalho em segundo plano (barramento entre réplicas, reprocessamento do spool).
func (m *Module) Run(ctx context.Context) error {
	m.api.hub.run(ctx)
	return nil
}

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

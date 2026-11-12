// Package identity reúne usuários, autenticação (senha e Google) e sessão.
// Substitui user-service, auth-service e auth-middleware.
//
// API para outros módulos: Directory (consulta de usuários) e authn (identidade na requisição).
// Todo o resto (SQL, handlers, OAuth) é privado ao pacote; nenhum outro módulo toca o schema
// "identity".
package identity

import (
	"context"
	"embed"
	"errors"
	"io/fs"
	"net/http"
	"time"

	"github.com/FelipeFelipeRenan/goverse/internal/platform/authn"
	"github.com/FelipeFelipeRenan/goverse/internal/platform/migrate"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

var (
	ErrNotFound   = errors.New("usuário não encontrado")
	ErrEmailTaken = errors.New("e-mail já cadastrado")
)

// User é a visão pública de um usuário (nunca contém a senha).
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Picture   string    `json:"picture"`
	IsOAuth   bool      `json:"is_oauth"`
	CreatedAt time.Time `json:"created_at"`
}

// Directory é o que os outros módulos (rooms, chat) enxergam de identity.
// Substitui o gRPC UserService: GetMany elimina o N+1 de GetRoomMembers.
type Directory interface {
	Exists(ctx context.Context, id string) (bool, error)
	Get(ctx context.Context, id string) (User, error)
	// ByEmail resolve um e-mail (normalizado) para o usuário; ErrNotFound se não existir.
	ByEmail(ctx context.Context, email string) (User, error)
	GetMany(ctx context.Context, ids []string) (map[string]User, error)
}

type Config struct {
	Secure             bool // true em produção (cookies Secure + SameSite=None)
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	RateLimit          bool // limita cadastro e login (por IP e, no login, por e-mail)
	TrustProxy         bool // IP do cliente via X-Forwarded-For (só atrás de proxy confiável)
}

type Module struct {
	pool *pgxpool.Pool
	st   store
	api  *api
}

func New(pool *pgxpool.Pool, keys *authn.Keys, cfg Config) *Module {
	st := &pgStore{pool: pool}
	return &Module{pool: pool, st: st, api: newAPI(st, keys, cfg)}
}

func (m *Module) Name() string { return "identity" }

func (m *Module) Register(mux *http.ServeMux) { m.api.routes(mux) }

func (m *Module) Migrate(ctx context.Context) error {
	sub, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		return err
	}
	return migrate.Run(ctx, m.pool, "identity", sub)
}

func (m *Module) Directory() Directory { return directory{m.st} }

type directory struct{ st store }

func (d directory) Exists(ctx context.Context, id string) (bool, error) {
	_, err := d.st.byID(ctx, id)
	if errors.Is(err, ErrNotFound) {
		return false, nil
	}
	return err == nil, err
}

func (d directory) Get(ctx context.Context, id string) (User, error) {
	r, err := d.st.byID(ctx, id)
	return r.User, err
}

func (d directory) ByEmail(ctx context.Context, email string) (User, error) {
	r, err := d.st.byEmail(ctx, normEmail(email))
	return r.User, err
}

func (d directory) GetMany(ctx context.Context, ids []string) (map[string]User, error) {
	users, err := d.st.many(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make(map[string]User, len(users))
	for _, u := range users {
		out[u.ID] = u
	}
	return out, nil
}

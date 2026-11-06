// Package rooms reúne salas e participantes. Substitui o room-service.
//
// API para outros módulos: Membership (o chat usa IsMember no lugar do gRPC antigo).
// Usuários são consultados por identity.Directory (chamada direta, sem rede).
package rooms

import (
	"context"
	"embed"
	"errors"
	"io/fs"
	"net/http"
	"time"

	"github.com/FelipeFelipeRenan/goverse/internal/identity"
	"github.com/FelipeFelipeRenan/goverse/internal/platform/authn"
	"github.com/FelipeFelipeRenan/goverse/internal/platform/migrate"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type Role string

const (
	RoleOwner     Role = "owner"
	RoleAdmin     Role = "admin"
	RoleModerator Role = "moderator"
	RoleMember    Role = "member"
	RoleGuest     Role = "guest"
)

func (r Role) valid() bool {
	switch r {
	case RoleOwner, RoleAdmin, RoleModerator, RoleMember, RoleGuest:
		return true
	}
	return false
}

var (
	ErrRoomNotFound        = errors.New("sala não encontrada")
	ErrMemberNotFound      = errors.New("membro não encontrado")
	ErrUserNotFound        = errors.New("usuário não encontrado")
	ErrMemberAlreadyExists = errors.New("usuário já é membro da sala")
	ErrRoomFull            = errors.New("sala lotada")
	ErrForbidden           = errors.New("ação não autorizada")
)

// invalidError vira 400 com a mensagem para o cliente.
type invalidError struct{ msg string }

func (e invalidError) Error() string { return e.msg }

func invalid(msg string) error { return invalidError{msg} }

type Room struct {
	ID          string
	Name        string
	Description string
	IsPublic    bool
	OwnerID     string
	MemberCount int
	MaxMembers  int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Member struct {
	RoomID   string
	UserID   string
	Role     Role
	JoinedAt time.Time
}

// Membership é o que os outros módulos (chat) enxergam de rooms.
type Membership interface {
	IsMember(ctx context.Context, roomID, userID string) (bool, error)
}

type Module struct {
	pool *pgxpool.Pool
	svc  *service
	api  *api
}

func New(pool *pgxpool.Pool, keys *authn.Keys, users identity.Directory) *Module {
	svc := &service{st: &pgStore{pool: pool}, users: users}
	return &Module{pool: pool, svc: svc, api: &api{svc: svc, keys: keys}}
}

func (m *Module) Name() string { return "rooms" }

func (m *Module) Register(mux *http.ServeMux) { m.api.routes(mux) }

func (m *Module) Migrate(ctx context.Context) error {
	sub, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		return err
	}
	return migrate.Run(ctx, m.pool, "rooms", sub)
}

func (m *Module) Membership() Membership { return m.svc }

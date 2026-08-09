package port

import (
	"context"

	"github.com/FelipeFelipeRenan/goverse/internal/modules/identity/domain"
)

// UserRepository defines the persistence operations required by the identity module.
type UserRepository interface {
	CreateUser(ctx context.Context, user domain.User) (*domain.UserResponse, error)
	UpdateUser(ctx context.Context, id string, user domain.User) (*domain.UserResponse, error)
	DeleteUser(ctx context.Context, id string) error
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	GetAllUsers(ctx context.Context) ([]domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
}

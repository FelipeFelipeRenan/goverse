package service

import (
	"context"

	"github.com/FelipeFelipeRenan/goverse/auth-service/internal/auth/domain"
)

type AuthMethod interface {
	Authenticate(ctx context.Context, credentials domain.Credentials) (*domain.UserResponse, error)
}

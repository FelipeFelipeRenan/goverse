package service

import (
	"context"

	"github.com/FelipeFelipeRenan/goverse/auth-service/internal/auth/domain"
	"github.com/FelipeFelipeRenan/goverse/auth-service/internal/auth/repository"
)

type PasswordAuth struct {
	authRepo repository.AuthRepository
}

func NewPasswordAuth(authRepo repository.AuthRepository) AuthMethod{
	return &PasswordAuth{authRepo}
}


func (a *PasswordAuth) Authenticate(ctx context.Context, credentials domain.Credentials) (*domain.UserResponse, error) {
	res, err := a.authRepo.ValidateCredentials(ctx, credentials.Email, credentials.Password)
	if err != nil {
		return nil, err
	}

	return &domain.UserResponse{
		ID:   res.Id,
		Username: res.Name,
	}, nil
}

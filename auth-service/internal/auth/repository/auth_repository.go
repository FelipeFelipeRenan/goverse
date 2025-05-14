package repository

import (
	"context"

	"github.com/FelipeFelipeRenan/goverse/auth-service/internal/auth/domain"
	userpb "github.com/FelipeFelipeRenan/goverse/proto/user"
)

type AuthRepository interface {
	ValidateCredentials(ctx context.Context, email, password string) (*userpb.UserResponse, error)
	FindByEmail(ctx context.Context, email string) (*domain.UserResponse, error)
}

type authRepository struct {
	userClient userpb.UserServiceClient
}

func NewAuthRepository(userClient userpb.UserServiceClient) AuthRepository {
	return &authRepository{userClient: userClient}
}

func (r *authRepository) ValidateCredentials(ctx context.Context, email, password string) (*userpb.UserResponse, error) {

	req := &userpb.CredentialsRequest{
		Email:    email,
		Password: password,
	}

	return r.userClient.ValidateCredentials(ctx, req)
}


func (r *authRepository) FindByEmail(ctx context.Context, email string) (*domain.UserResponse, error){
	req := &userpb.EmailRequest{
		Email: email,
	}

	resp, err := r.userClient.GetUserByEmail(ctx, req)
	if err != nil {
		return nil, err
	}

	return &domain.UserResponse{
		ID:       resp.Id,
		Username: resp.Name,
		Email:    resp.Email,
	}, nil
}
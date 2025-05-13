package repository

import (
	"context"

	userpb "github.com/FelipeFelipeRenan/goverse/proto/user"
)

type AuthRepository interface {
	ValidateCredentials(ctx context.Context, email, password string) (*userpb.UserResponse, error)
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

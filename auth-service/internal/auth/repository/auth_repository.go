package repository

import (
	"context"
	"fmt"

	"github.com/FelipeFelipeRenan/goverse/auth-service/internal/auth/domain"
	"github.com/FelipeFelipeRenan/goverse/proto/user"
	"github.com/jackc/pgx/v5"
)


type AuthRepository interface {
	GetUserByEmail(ctx context.Context, email string)(*user.User, error )
}


type authRepository struct {
  userClient user.UserServiceClient
}

func NewAuthRepository(userClient user.UserServiceClient) AuthRepository{
	return &authRepository{userClient:userClient}
}

func (r *authRepository) GetUserByEmail(ctx context.Context, email string)(*user.User, error){
	req := &user.GetUserByEmailRequest{Email: email}
	resp, err := r.userClient.GetUserByEmail(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("erro na chamada do grpc")
	}
	return resp.User, nil
}
package client

import (
	"context"
	"fmt"

	userpb "github.com/FelipeFelipeRenan/goverse/proto/user"
	"google.golang.org/grpc"
)

type UserServiceClient interface {
	ExistsUserByID(ctx context.Context, userID string) (bool, error)
	IsUserValid(ctx context.Context, userID string) (bool, error)
	GetUserByID(ctx context.Context, userID string) (*userpb.UserResponse, error)
}

type userServiceClient struct {
	grpcClient userpb.UserServiceClient
}

func NewUserServiceClient(cc *grpc.ClientConn) UserServiceClient {
	return &userServiceClient{
		grpcClient: userpb.NewUserServiceClient(cc),
	}
}

// ExistsUserByID implements UserServiceClient.
func (u *userServiceClient) ExistsUserByID(ctx context.Context, userID string) (bool, error) {
	resp, err := u.grpcClient.ExistsUserByID(ctx, &userpb.UserIDRequest{Id: userID})
	if err != nil {
		return false, fmt.Errorf("erro ao verificar existencia do usuário: %w", err)
	}
	return resp.GetExists(), nil
}

func (u *userServiceClient) IsUserValid(ctx context.Context, userID string) (bool, error) {
	return u.ExistsUserByID(ctx, userID)
}

func (u *userServiceClient) GetUserByID(ctx context.Context, userID string) (*userpb.UserResponse, error) {
	resp, err := u.grpcClient.GetUserByID(ctx, &userpb.UserIDRequest{Id: userID})
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar usuário por ID: %w", err)
	}
	return resp, nil
}

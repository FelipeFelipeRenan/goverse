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

func (c *userServiceClient) IsUserValid(ctx context.Context, userID string) (bool, error) {
	return c.ExistsUserByID(ctx, userID)
}

package mocks

import (
	"context"

	userpb "github.com/FelipeFelipeRenan/goverse/proto/user"
	"github.com/stretchr/testify/mock"
)

// MockUserServiceClient é um mock da interface UserServiceClient
type MockUserServiceClient struct {
	mock.Mock
}

// GetUserByID implements client.UserServiceClient.
func (m *MockUserServiceClient) GetUserByID(ctx context.Context, userID string) (*userpb.UserResponse, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*userpb.UserResponse), args.Error(1)
}

// IsUserValid implements client.UserServiceClient.
func (m *MockUserServiceClient) IsUserValid(ctx context.Context, userID string) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

// ExistsUserByID implements client.UserServiceClient.
func (m *MockUserServiceClient) ExistsUserByID(ctx context.Context, userID string) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

// func (m *MockUserServiceClient) GetUserByID(ctx context.Context, userID string) (*domain.RoomMember, error) {
// 	args := m.Called(ctx, userID)
// 	return args.Get(0).(*domain.RoomMember), args.Error(1)
// }

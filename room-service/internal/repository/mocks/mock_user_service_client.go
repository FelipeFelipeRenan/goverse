package mocks

import (
	"context"

	"github.com/FelipeFelipeRenan/goverse/room-service/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockUserServiceClient é um mock da interface UserServiceClient
type MockUserServiceClient struct {
	mock.Mock
}

// ExistsUserByID implements client.UserServiceClient.
func (m *MockUserServiceClient) ExistsUserByID(ctx context.Context, userID string) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserServiceClient) GetUserByID(ctx context.Context, userID string) (*domain.RoomMember, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*domain.RoomMember), args.Error(1)
}

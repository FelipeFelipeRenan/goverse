package repository

import (
	"context"

	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/domain"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

// GetUserByEmail implements UserRepository.
func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	panic("unimplemented")
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user domain.User) (*domain.UserResponse, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(*domain.UserResponse), args.Error(1)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetAllUsers(ctx context.Context) ([]domain.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.User), args.Error(1)
}

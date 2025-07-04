package service

import (
	"context"

	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/domain"
	"github.com/stretchr/testify/mock"
)

type MockUserService struct {
	mock.Mock
}

// DeleteUser implements UserService.
func (m *MockUserService) DeleteUser(ctx context.Context, id string) error {
	panic("unimplemented")
}

// UpdateUser implements UserService.
func (m *MockUserService) UpdateUser(ctx context.Context, id string, user domain.User) (*domain.UserResponse, error) {
	panic("unimplemented")
}

// ExistsByID implements UserService.
func (m *MockUserService) ExistsByID(ctx context.Context, id string) (bool, error) {
	panic("unimplemented")
}

// GetByEmail implements UserService.
func (m *MockUserService) GetByEmail(ctx context.Context, email string) (*domain.UserResponse, error) {
	panic("unimplemented")
}

// Authenticate implements UserService.
func (m *MockUserService) Authenticate(ctx context.Context, email, password string) (*domain.User, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) Register(ctx context.Context, user domain.User) (*domain.UserResponse, error) {
	args := m.Called(ctx, user)
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.(*domain.UserResponse), args.Error(1)
}

func (m *MockUserService) FindByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) GetAllUsers(ctx context.Context) ([]domain.User, error) {
	args := m.Called(ctx)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]domain.User), args.Error(1)
}

func (m *MockUserService) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)

}

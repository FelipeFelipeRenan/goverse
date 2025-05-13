package service

import (
	"context"

	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/domain"
	"github.com/stretchr/testify/mock"
)

type MockUserService struct {
	mock.Mock
}

// Authenticate implements UserService.
func (m *MockUserService) Authenticate(ctx context.Context, email string, password string) (*domain.User, error) {
	panic("unimplemented")
}

func (m *MockUserService) Register(ctx context.Context, user domain.User) (string, error) {
	args := m.Called(ctx, user)
	return args.String(0), args.Error(1)
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

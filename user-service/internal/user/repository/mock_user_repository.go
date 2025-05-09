package repository

import (
	"context"

	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/domain"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}


func (m *MockUserRepository) CreateUser(ctx context.Context, user domain.User) (string, error){
	args := m.Called(ctx, user)
	return args.String(0), args.Error(1)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id string)(*domain.User, error){
	args := m.Called(ctx, id)
	if args.Get(0) == nil{
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetAllUsers(ctx context.Context) ([]domain.User, error){
	args := m.Called(ctx)
	return args.Get(0).([]domain.User), args.Error(1)
}
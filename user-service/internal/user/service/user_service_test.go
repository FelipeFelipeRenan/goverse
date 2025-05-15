package service

import (
	"context"
	"errors"
	"testing"

	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/domain"
	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserService_Register_Success(t *testing.T) {
	t.Parallel()
	mockRepo := new(repository.MockUserRepository)
	svc := NewUserService(mockRepo)

	user := domain.User{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "securepass",
	}

	expectedResp := &domain.UserResponse{ID: "123"}

	mockRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("domain.User")).Return(expectedResp, nil)
	id, err := svc.Register(context.Background(), user)

	assert.NoError(t, err)
	assert.Equal(t, "123", id)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Register_InvalidData(t *testing.T) {
	t.Parallel()
	mockRepo := new(repository.MockUserRepository)
	svc := NewUserService(mockRepo)

	user := domain.User{} // Dados vazios

	id, err := svc.Register(context.Background(), user)

	assert.Error(t, err)
	assert.Equal(t, "", id)
}

func TestUserService_FindByID_Success(t *testing.T) {
	t.Parallel()
	mockRepo := new(repository.MockUserRepository)
	svc := NewUserService(mockRepo)

	expected := &domain.User{
		ID:       "123",
		Username: "bob",
		Email:    "bob@example.com",
	}

	mockRepo.On("GetUserByID", mock.Anything, "123").Return(expected, nil)

	user, err := svc.FindByID(context.Background(), "123")

	assert.NoError(t, err)
	assert.Equal(t, expected, user)
	mockRepo.AssertExpectations(t)
}

func TestUserService_FindByID_NotFound(t *testing.T) {
	t.Parallel()
	mockRepo := new(repository.MockUserRepository)
	svc := NewUserService(mockRepo)

	mockRepo.On("GetUserByID", mock.Anything, "not_found").Return(nil, errors.New("not found"))

	user, err := svc.FindByID(context.Background(), "not_found")

	assert.Error(t, err)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetAllUsers(t *testing.T) {
	t.Parallel()
	mockRepo := new(repository.MockUserRepository)
	svc := NewUserService(mockRepo)
	expectedUsers := []domain.User{
		{ID: "1", Username: "alice", Email: "alice@example.com"},
		{ID: "2", Username: "bob", Email: "bob@example.com"},
	}

	mockRepo.On("GetAllUsers", mock.Anything).Return(expectedUsers, nil)

	users, err := svc.GetAllUsers(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, expectedUsers, users)
	mockRepo.AssertExpectations(t)
}

package mocks

import (
	"context"

	"github.com/FelipeFelipeRenan/goverse/room-service/internal/domain"
	"github.com/FelipeFelipeRenan/goverse/room-service/internal/repository"
	"github.com/stretchr/testify/mock"
)

type MockRoomRepository struct {
	mock.Mock
}

func (m *MockRoomRepository) Create(ctx context.Context, dbtx repository.DBTX, room *domain.Room) error {
	args := m.Called(ctx, dbtx, room)
	return args.Error(0)
}

func (m *MockRoomRepository) GetByID(ctx context.Context, dbtx repository.DBTX, id string) (*domain.Room, error) {
	args := m.Called(ctx, dbtx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Room), args.Error(1)
}

func (m *MockRoomRepository) ListAll(ctx context.Context, dbtx repository.DBTX, limit, offset int) ([]*domain.Room, error) {
	args := m.Called(ctx, dbtx, limit, offset)
	return args.Get(0).([]*domain.Room), args.Error(1)
}

func (m *MockRoomRepository) ListPublic(ctx context.Context, dbtx repository.DBTX, limit, offset int) ([]*domain.Room, error) {
	args := m.Called(ctx, dbtx, limit, offset)
	return args.Get(0).([]*domain.Room), args.Error(1)
}

func (m *MockRoomRepository) Update(ctx context.Context, dbtx repository.DBTX, room *domain.Room) error {
	args := m.Called(ctx, dbtx, room)
	return args.Error(0)
}

func (m *MockRoomRepository) Delete(ctx context.Context, dbtx repository.DBTX, id string) error {
	args := m.Called(ctx, dbtx, id)
	return args.Error(0)
}

func (m *MockRoomRepository) SearchByName(ctx context.Context, dbtx repository.DBTX, keyword string) ([]*domain.Room, error) {
	args := m.Called(ctx, dbtx, keyword)
	return args.Get(0).([]*domain.Room), args.Error(1)
}

func (m *MockRoomRepository) IncrementMemberCount(ctx context.Context, dbtx repository.DBTX, roomID string, delta int) error {
	args := m.Called(ctx, dbtx, roomID, delta)
	return args.Error(0)
}

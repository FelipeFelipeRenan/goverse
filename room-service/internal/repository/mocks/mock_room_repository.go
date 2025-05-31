package mocks

import (
	"context"

	"github.com/FelipeFelipeRenan/goverse/room-service/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockRoomRepository struct {
	mock.Mock
}

func (m *MockRoomRepository) Create(ctx context.Context, room *domain.Room) error {
	args := m.Called(ctx, room)
	return args.Error(0)
}

func (m *MockRoomRepository) GetByID(ctx context.Context, id string) (*domain.Room, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Room), args.Error(1)
}

func (m *MockRoomRepository) ListPublic(ctx context.Context) ([]*domain.Room, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*domain.Room), args.Error(1)
}

func (m *MockRoomRepository) ListByUserID(ctx context.Context, userID string) ([]*domain.Room, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*domain.Room), args.Error(1)
}

func (m *MockRoomRepository) Update(ctx context.Context, room *domain.Room) error {
	args := m.Called(ctx, room)
	return args.Error(0)
}

func (m *MockRoomRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Exists implements repository.RoomRepository.
func (m *MockRoomRepository) Exists(ctx context.Context, id string) (bool, error) {
	panic("unimplemented")
}

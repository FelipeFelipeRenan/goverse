package mocks

import (
	"context"

	"github.com/FelipeFelipeRenan/goverse/room-service/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockRoomRepository struct {
	mock.Mock
}

// AddMember implements repository.RoomMemberRepository.
func (m *MockRoomRepository) AddMember(ctx context.Context, member *domain.RoomMember) error {
	panic("unimplemented")
}

// GetMemberByID implements repository.RoomMemberRepository.
func (m *MockRoomRepository) GetMemberByID(ctx context.Context, roomID string, userID string) (*domain.RoomMember, error) {
	panic("unimplemented")
}

// GetMemberRole implements repository.RoomMemberRepository.
func (m *MockRoomRepository) GetMemberRole(ctx context.Context, roomID string, userID string) (domain.Role, error) {
	panic("unimplemented")
}

// GetMembers implements repository.RoomMemberRepository.
func (m *MockRoomRepository) GetMembers(ctx context.Context, roomID string) ([]*domain.RoomMember, error) {
	panic("unimplemented")
}

// IsMember implements repository.RoomMemberRepository.
func (m *MockRoomRepository) IsMember(ctx context.Context, roomID string, userID string) (bool, error) {
	panic("unimplemented")
}

// RemoveMember implements repository.RoomMemberRepository.
func (m *MockRoomRepository) RemoveMember(ctx context.Context, roomID string, userID string) error {
	panic("unimplemented")
}

// UpdateMemberRole implements repository.RoomMemberRepository.
func (m *MockRoomRepository) UpdateMemberRole(ctx context.Context, roomID string, userID string, newRole domain.Role) error {
	panic("unimplemented")
}

// ListAll implements repository.RoomRepository.
func (m *MockRoomRepository) ListAll(ctx context.Context, limit int, offset int) ([]*domain.Room, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*domain.Room), args.Error(1)
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

func (m *MockRoomRepository) ListPublic(ctx context.Context, limit, offset int) ([]*domain.Room, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*domain.Room), args.Error(1)
}

func (m *MockRoomRepository) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.Room, error) {
	args := m.Called(ctx, userID, limit, offset)
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

func (m *MockRoomRepository) Exists(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockRoomRepository) IncrementMemberCount(ctx context.Context, roomID string, delta int) error {
	args := m.Called(ctx, roomID, delta)
	return args.Error(0)
}

func (m *MockRoomRepository) SearchByName(ctx context.Context, keyword string) ([]*domain.Room, error) {
	args := m.Called(ctx, keyword)
	return args.Get(0).([]*domain.Room), args.Error(1)
}

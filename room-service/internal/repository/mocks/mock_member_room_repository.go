package mocks

import (
	"context"

	"github.com/FelipeFelipeRenan/goverse/room-service/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockRoomMemberRepository struct {
	mock.Mock
}

// GetRoomsByOwnerID implements repository.RoomMemberRepository.
func (m *MockRoomMemberRepository) GetRoomsByOwnerID(ctx context.Context, userID string) ([]*domain.Room, error) {
	panic("unimplemented")
}

// GetRoomsByUserID implements repository.RoomMemberRepository.
func (m *MockRoomMemberRepository) GetRoomsByUserID(ctx context.Context, userID string) ([]*domain.Room, error) {
	panic("unimplemented")
}

func (m *MockRoomMemberRepository) AddMember(ctx context.Context, member *domain.RoomMember) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

func (m *MockRoomMemberRepository) RemoveMember(ctx context.Context, roomID, userID string) error {
	args := m.Called(ctx, roomID, userID)
	return args.Error(0)
}

func (m *MockRoomMemberRepository) GetMembers(ctx context.Context, roomID string) ([]*domain.RoomMember, error) {
	args := m.Called(ctx, roomID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.RoomMember), args.Error(1)
}

func (m *MockRoomMemberRepository) GetMemberByID(ctx context.Context, roomID, userID string) (*domain.RoomMember, error) {
	args := m.Called(ctx, roomID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RoomMember), args.Error(1)
}

func (m *MockRoomMemberRepository) GetMemberRole(ctx context.Context, roomID, userID string) (domain.Role, error) {
	args := m.Called(ctx, roomID, userID)
	return args.Get(0).(domain.Role), args.Error(1)
}

func (m *MockRoomMemberRepository) UpdateMemberRole(ctx context.Context, roomID, userID string, newRole domain.Role) error {
	args := m.Called(ctx, roomID, userID, newRole)
	return args.Error(0)
}

func (m *MockRoomMemberRepository) IsMember(ctx context.Context, roomID, userID string) (bool, error) {
	args := m.Called(ctx, roomID, userID)
	return args.Bool(0), args.Error(1)
}

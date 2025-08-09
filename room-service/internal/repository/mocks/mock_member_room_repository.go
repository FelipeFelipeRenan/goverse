package mocks

import (
	"context"

	"github.com/FelipeFelipeRenan/goverse/room-service/internal/domain"
	"github.com/FelipeFelipeRenan/goverse/room-service/internal/repository"
	"github.com/stretchr/testify/mock"
)

type MockRoomMemberRepository struct {
	mock.Mock
}

func (m *MockRoomMemberRepository) AddMember(ctx context.Context, dbtx repository.DBTX, member *domain.RoomMember) error {
	args := m.Called(ctx, dbtx, member)
	return args.Error(0)
}

func (m *MockRoomMemberRepository) RemoveMember(ctx context.Context, dbtx repository.DBTX, roomID, userID string) error {
	args := m.Called(ctx, dbtx, roomID, userID)
	return args.Error(0)
}

func (m *MockRoomMemberRepository) GetMembers(ctx context.Context, dbtx repository.DBTX, roomID string) ([]*domain.RoomMember, error) {
	args := m.Called(ctx, dbtx, roomID)
	return args.Get(0).([]*domain.RoomMember), args.Error(1)
}

func (m *MockRoomMemberRepository) GetMemberByID(ctx context.Context, dbtx repository.DBTX, roomID, userID string) (*domain.RoomMember, error) {
	args := m.Called(ctx, dbtx, roomID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RoomMember), args.Error(1)
}

func (m *MockRoomMemberRepository) UpdateMemberRole(ctx context.Context, dbtx repository.DBTX, roomID, userID string, newRole domain.Role) error {
	args := m.Called(ctx, dbtx, roomID, userID, newRole)
	return args.Error(0)
}

func (m *MockRoomMemberRepository) GetRoomsByUserID(ctx context.Context, dbtx repository.DBTX, userID string) ([]*domain.Room, error) {
	args := m.Called(ctx, dbtx, userID)
	return args.Get(0).([]*domain.Room), args.Error(1)
}

func (m *MockRoomMemberRepository) GetRoomsByOwnerID(ctx context.Context, dbtx repository.DBTX, userID string) ([]*domain.Room, error) {
	args := m.Called(ctx, dbtx, userID)
	return args.Get(0).([]*domain.Room), args.Error(1)
}

func (m *MockRoomMemberRepository) IsMember(ctx context.Context, dbtx repository.DBTX, roomID, userID string) (bool, error) {
	args := m.Called(ctx, dbtx, roomID, userID)
	return args.Bool(0), args.Error(1)
}

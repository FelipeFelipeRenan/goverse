package service_test

import (
	"context"
	"testing"

	"github.com/FelipeFelipeRenan/goverse/room-service/internal/domain"
	"github.com/FelipeFelipeRenan/goverse/room-service/internal/repository/mocks"
	"github.com/FelipeFelipeRenan/goverse/room-service/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateRoom_Success(t *testing.T) {
	t.Parallel()
	
	roomRepo := new(mocks.MockRoomRepository)
	memberRepo := new(mocks.MockRoomMemberRepository)

	roomService := service.NewRoomService(roomRepo, memberRepo)

	room := &domain.Room{
		ID:          "1",
		Name:        "Sala de Teste",
		Description: "Uma sala de teste",
		IsPublic:    true,
		OwnerID:     "123",
	}

	roomRepo.On("Create", mock.Anything, room).Return(nil)
	memberRepo.On("AddMember", mock.Anything, mock.MatchedBy(func(m *domain.RoomMember) bool {
		return m.UserID == "123" && m.Role == domain.RoleOwner
	})).Return(nil)

	_, err := roomService.CreateRoom(context.Background(), room.OwnerID, room)

	assert.NoError(t, err)
	roomRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)

}

func TestRoomService_AddMember_Success(t *testing.T) {
	t.Parallel()

	roomRepo := new(mocks.MockRoomRepository)
	memberRepo := new(mocks.MockRoomMemberRepository)

	roomService := service.NewRoomService(roomRepo, memberRepo)

	ctx := context.Background()
	roomID := "1"
	userID := "2"
	actorID := "999"
	role := domain.RoleMember

	memberRepo.On("IsMember", ctx, roomID, userID).Return(false, nil)
	memberRepo.On("AddMember", ctx, mock.MatchedBy(func(member *domain.RoomMember) bool {
		return member.RoomID == roomID && member.UserID == userID && member.Role == role
	})).Return(nil)

	err := roomService.AddMember(ctx, actorID, roomID, userID, role)

	require.NoError(t, err)
	memberRepo.AssertExpectations(t)
}

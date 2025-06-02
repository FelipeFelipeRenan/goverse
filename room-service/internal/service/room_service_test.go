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
	userClient := new(mocks.MockUserServiceClient)

	memberService := service.NewMemberService(memberRepo, roomRepo, userClient)

	ctx := context.Background()
	roomID := "1"
	userID := "2"
	actorID := "999" // deve ser OwnerID da sala para passar autorização
	role := domain.RoleMember

	// Mocka retorno da sala com OwnerID == actorID para autorização
	roomRepo.On("GetByID", ctx, roomID).Return(&domain.Room{
		ID:      roomID,
		OwnerID: actorID,
	}, nil)

	// Mocka consulta de existência de usuário via gRPC
	userClient.On("ExistsUserByID", ctx, userID).Return(true, nil)

	// Mocka verificação de que o usuário ainda não é membro
	memberRepo.On("GetMemberByID", ctx, roomID, userID).Return(nil, domain.ErrMemberNotFound)

	// Mocka inserção do novo membro
	memberRepo.On("AddMember", ctx, mock.MatchedBy(func(member *domain.RoomMember) bool {
		return member.RoomID == roomID && member.UserID == userID && member.Role == role
	})).Return(nil)

	// Mocka incremento do contador de membros
	roomRepo.On("IncrementMemberCount", ctx, roomID, 1).Return(nil)

	err := memberService.AddMember(ctx, actorID, roomID, userID, role)

	require.NoError(t, err)
	memberRepo.AssertExpectations(t)
	roomRepo.AssertExpectations(t)
	userClient.AssertExpectations(t)
}

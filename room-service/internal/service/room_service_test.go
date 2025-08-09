package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/FelipeFelipeRenan/goverse/room-service/internal/domain"
	"github.com/FelipeFelipeRenan/goverse/room-service/internal/repository/mocks"
	"github.com/FelipeFelipeRenan/goverse/room-service/internal/service"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock para a nossa interface DBPool.
// Como DBPool agora inclui DBTX, precisamos mockar todos os métodos.
// Na prática, só precisamos definir o comportamento dos que usamos no teste.
type MockDBPool struct {
	mock.Mock
}

// Implementação do mock para a interface DBPool
func (m *MockDBPool) Begin(ctx context.Context) (pgx.Tx, error) {
	args := m.Called(ctx)
	// A assinatura agora está correta, retornando a interface pgx.Tx
	return args.Get(0).(pgx.Tx), args.Error(1)
}

// Métodos da interface DBTX herdada (não precisamos deles neste teste, mas precisam existir)
func (m *MockDBPool) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	args := m.Called(ctx, sql, arguments)
	return args.Get(0).(pgconn.CommandTag), args.Error(1)
}
func (m *MockDBPool) Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error) {
	args := m.Called(ctx, sql, arguments)
	return args.Get(0).(pgx.Rows), args.Error(1)
}
func (m *MockDBPool) QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row {
	args := m.Called(ctx, sql, arguments)
	return args.Get(0).(pgx.Row)
}

func TestCreateRoom_Success(t *testing.T) {
	// Setup
	mockPool := new(MockDBPool)
	mockRoomRepo := new(mocks.MockRoomRepository)
	mockMemberRepo := new(mocks.MockRoomMemberRepository)
	roomService := service.NewRoomService(mockPool, mockRoomRepo, mockMemberRepo)

	room := &domain.Room{
		Name:    "Sala de Teste",
		OwnerID: "123",
	}

	mockTx, err := pgxmock.NewConn() // Cria um mock de conexão que pode agir como Tx
	require.NoError(t, err)

	// Expectativas
	mockPool.On("Begin", mock.Anything).Return(pgx.Tx(mockTx), nil) // Faz o cast para a interface
	mockRoomRepo.On("Create", mock.Anything, pgx.Tx(mockTx), room).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(2).(*domain.Room)
		arg.ID = "1" // Simula o retorno do ID pelo banco
	})
	mockMemberRepo.On("AddMember", mock.Anything, pgx.Tx(mockTx), mock.AnythingOfType("*domain.RoomMember")).Return(nil)
	mockRoomRepo.On("IncrementMemberCount", mock.Anything, pgx.Tx(mockTx), "1", 1).Return(nil)
	mockTx.ExpectCommit()

	// Execução
	createdRoom, err := roomService.CreateRoom(context.Background(), "123", room)

	// Verificação
	require.NoError(t, err)
	require.Equal(t, "1", createdRoom.ID)
	require.Equal(t, 1, createdRoom.MemberCount)

	mockPool.AssertExpectations(t)
	mockRoomRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
	require.NoError(t, mockTx.ExpectationsWereMet())
}

func TestCreateRoom_Failure_ShouldRollback(t *testing.T) {
	mockPool := new(MockDBPool)
	mockRoomRepo := new(mocks.MockRoomRepository)
	mockMemberRepo := new(mocks.MockRoomMemberRepository)
	roomService := service.NewRoomService(mockPool, mockRoomRepo, mockMemberRepo)

	room := &domain.Room{Name: "Sala de Teste", OwnerID: "123"}
	mockTx, err := pgxmock.NewConn()
	require.NoError(t, err)

	// Expectativas
	mockPool.On("Begin", mock.Anything).Return(pgx.Tx(mockTx), nil)
	mockRoomRepo.On("Create", mock.Anything, pgx.Tx(mockTx), room).Return(nil)
	mockMemberRepo.On("AddMember", mock.Anything, pgx.Tx(mockTx), mock.AnythingOfType("*domain.RoomMember")).Return(errors.New("db error"))
	mockTx.ExpectRollback()

	// Execução
	_, err = roomService.CreateRoom(context.Background(), "123", room)

	// Verificação
	require.Error(t, err)
	mockPool.AssertExpectations(t)
	mockRoomRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
	require.NoError(t, mockTx.ExpectationsWereMet())
}

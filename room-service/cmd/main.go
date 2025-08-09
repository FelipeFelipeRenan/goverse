package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	//userpb "github.com/FelipeFelipeRenan/goverse/proto/user"
	"github.com/FelipeFelipeRenan/goverse/room-service/internal/client"
	"github.com/FelipeFelipeRenan/goverse/room-service/internal/delivery/routes"
	"github.com/FelipeFelipeRenan/goverse/room-service/internal/handler"
	"github.com/FelipeFelipeRenan/goverse/room-service/internal/repository"
	"github.com/FelipeFelipeRenan/goverse/room-service/internal/service"
	"github.com/FelipeFelipeRenan/goverse/room-service/pkg/database"
	"github.com/FelipeFelipeRenan/goverse/room-service/pkg/logger"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {

	if os.Getenv("ENV") != "prod" {
		erro := godotenv.Load()
		if erro != nil {
			logger.Error("Erro ao carregar .env", "err", erro)
		}
	}

	logger.Init("info", "room-service")

	// Conexão com banco de dados
	dbPool, err := database.Connect()
	if err != nil {
		logger.Error("Erro ao conectar ao banco de dados", "err", err)
		return
	}
	defer dbPool.Close()

	if err := database.RunMigration(dbPool); err != nil {
		logger.Error("Erro ao rodar migrações", "err", err)
		return
	}

	grpc_host := os.Getenv("GRPC_SERVER_HOST")
	grpc_port := os.Getenv("GRPC_SERVER_PORT")
	// Conexão com user-service via gRPC
	conn, err := grpc.NewClient(grpc_host+grpc_port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("falha ao conectar ao user-service: ", "err", err)
	}
	defer conn.Close()
	userClient := client.NewUserServiceClient(conn)

	// Inicializa repositórios e serviços
	roomRepo := repository.NewRoomRepository()
	memberRepo := repository.NewRoomMemberRepository()
	roomService := service.NewRoomService(dbPool, roomRepo, memberRepo)
	memberService := service.NewMemberService(dbPool, memberRepo, roomRepo, userClient)

	// Inicializa handlers e rotas
	roomHandler := handler.NewRoomHandler(roomService, userClient)
	memberHandler := handler.NewMemberHandler(memberService)
	routes.RegisterRoutes(roomHandler, memberHandler)

	port := os.Getenv("ROOM_SERVICE_PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Graceful shutdown
	go func() {
		logger.Info("Serviço de salas rodando", "port", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Erro ao iniciar servidor HTTP", "err", err)
		}
	}()

	// Espera sinal de encerramento (CTRL+C ou kill)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Info("Encerrando room service...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Erro ao encerrar servidor HTTP", "err", err)
	}
}

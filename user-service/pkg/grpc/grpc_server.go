package grpc

import (
	"net"
	"os"

	userpb "github.com/FelipeFelipeRenan/goverse/proto/user"
	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/handler"
	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/service"
	"github.com/FelipeFelipeRenan/goverse/user-service/pkg/logger"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func StartGRPCServer(userService service.UserService) {

	if os.Getenv("ENV") != "prod" {
		erro := godotenv.Load()
		if erro != nil {
			logger.Error("Erro ao carregar .env", "err", erro)
		}
	}

	grpc_port := os.Getenv("GRPC_PORT")
	listener, err := net.Listen("tcp", grpc_port)
	if err != nil {
		logger.Error("Erro ao iniciar o listener gRPC", "err", err)
	}

	grpcServer := grpc.NewServer()

	userpb.RegisterUserServiceServer(grpcServer, handler.NewGRPCHandler(userService))

	logger.Info("Servidor gRPC ouvindo", "port", grpc_port)

	if err := grpcServer.Serve(listener); err != nil {
		logger.Error("Erro ao iniciar o servidor gRPC", "err", err)
	}
}

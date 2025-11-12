package grpc

import (
	"net"
	"os"

	roompb "github.com/FelipeFelipeRenan/goverse/proto/room"
	"github.com/FelipeFelipeRenan/goverse/room-service/internal/handler"
	"github.com/FelipeFelipeRenan/goverse/room-service/pkg/logger"
	"google.golang.org/grpc"
)

func StartGRPCServer(grpcHandler *handler.GRPCHandler) {
	grpc_port := os.Getenv("GRPC_PORT")
	if grpc_port == "" {
		grpc_port = ":50052"
	}

	listener, err := net.Listen("tcp", grpc_port)
	if err != nil {
		logger.Error("Erro ao iniciar o listener gRPC", "err", err)
		return
	}

	grpcServer := grpc.NewServer()
	roompb.RegisterRoomServiceServer(grpcServer, grpcHandler)

	logger.Info("Servidor gRPC (Room Service) ouvindo", "port", grpc_port)

	if err := grpcServer.Serve(listener); err != nil {
		logger.Error("Erro ao iniciar o servidor gRPC", "err", err)
	}
}

package grpc

import (
	"log"
	"net"
	"os"

	userpb "github.com/FelipeFelipeRenan/goverse/proto/user"
	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/handler"
	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/service"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)


func StartGRPCServer(userService service.UserService){

	godotenv.Load(".env")
	grpc_port := os.Getenv("GRPC_PORT")
	listener, err := net.Listen("tcp",grpc_port)
	if err != nil {
		log.Fatalf("Erro ao iniciar o listener gRPC: %v", err)
	}

	grpcServer := grpc.NewServer()

	userpb.RegisterUserServiceServer(grpcServer, handler.NewGRPCHandler(userService))

	log.Println("Servidor gRPC ouvindo na porta: " , grpc_port)
	
	if err := grpcServer.Serve(listener); err != nil{
		log.Fatalf("Erro ao iniciar o servidor gRPC: %v", err)
	}
}
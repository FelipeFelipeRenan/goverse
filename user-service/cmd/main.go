package main

import (
	"net/http"
	"os"

	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/delivery/rest/routes"
	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/handler"
	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/repository"
	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/service"
	"github.com/FelipeFelipeRenan/goverse/user-service/pkg/database"
	"github.com/FelipeFelipeRenan/goverse/user-service/pkg/grpc"
	"github.com/FelipeFelipeRenan/goverse/user-service/pkg/logger"
	"github.com/joho/godotenv"
)

func main() {
	erro := godotenv.Load(".env")

	logger.Init("info", "room-service")

	if erro != nil{
		logger.Error("Erro ao carregar .env", "err", erro)
	}
	
	conn, err := database.Connect()
	if err != nil {
		logger.Error("Erro ao conectar com banco de dados", "err", err)
	}
	
	defer conn.Close(nil)

	database.RunMigration(conn)

	repo := repository.NewUserRepository(conn)
	userService := service.NewUserService(repo)
	userHandler := handler.NewUserHandler(userService)

	// Goroutine para iniciar o servidor HTTP
	go func() {
		routes.SetupUserRoutes(userHandler)
		// Iniciando o servidor na porta 8080
		port := os.Getenv("APP_PORT")
		logger.Info("Serviço de usuários rodando", "port", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			logger.Error("Erro ao iniciar o serviço de usuários", "err", err)
		}
	}()

	// Go routine para iniciar servidor gRPC
	go grpc.StartGRPCServer(userService)

	select {}
}

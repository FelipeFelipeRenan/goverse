package main

import (
	"net/http"
	"os"

	"github.com/FelipeFelipeRenan/goverse/common/pkg/config"
	"github.com/FelipeFelipeRenan/goverse/common/pkg/database"
	"github.com/FelipeFelipeRenan/goverse/common/pkg/logger"
	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/delivery/rest/routes"
	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/handler"
	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/repository"
	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/service"
	"github.com/FelipeFelipeRenan/goverse/user-service/pkg/grpc"
)

func main() {

	logger.Init("info", "user-service")

	// Fail Fast: Se não tiver essas variáveis, nem tenta subir
	if err := config.RequireEnv("DB_HOST", "DB_USER", "GRPC_PORT"); err != nil {
		logger.Error("Erro de configuração inicial", "err", err)
		os.Exit(1)
	}

	conn, err := database.Connect()
	if err != nil {
		logger.Error("Erro ao conectar com banco de dados", "err", err)
	}

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

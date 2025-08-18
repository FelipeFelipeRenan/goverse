package main

import (
	"net/http"
	"os"

	"github.com/FelipeFelipeRenan/goverse/auth-service/internal/auth/delivery/rest/routes"
	"github.com/FelipeFelipeRenan/goverse/auth-service/internal/auth/handler"
	"github.com/FelipeFelipeRenan/goverse/auth-service/internal/auth/repository"
	"github.com/FelipeFelipeRenan/goverse/auth-service/internal/auth/service"
	"github.com/FelipeFelipeRenan/goverse/auth-service/pkg/logger"
	userpb "github.com/FelipeFelipeRenan/goverse/proto/user"
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

	logger.Init("info", "auth-service")

	privateKeyBytes, err := os.ReadFile(os.Getenv("JWT_PRIVATE_KEY_PATH"))
	if err != nil {
		logger.Error("falha ao ler a chave privada", "err", err)
	}
	grpc_host := os.Getenv("GRPC_SERVER_HOST")
	grpc_port := os.Getenv("GRPC_SERVER_PORT")
	conn, err := grpc.NewClient(grpc_host+grpc_port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("falha ao conectar ao user-service: ", "err", err)
	}
	defer conn.Close()

	userClient := userpb.NewUserServiceClient(conn)

	authRepo := repository.NewAuthRepository(userClient)

	passwordAuth := service.NewPasswordAuth(authRepo)

	authMethods := map[string]service.AuthMethod{
		"password": passwordAuth,
		"oauth":    service.NewOAuthAuth(authRepo),
	}
	// jwt_secret := os.Getenv("JWT_SECRET")
	authService := service.NewAuthService(authMethods, authRepo, privateKeyBytes)

	authHandler := handler.NewAuthHandler(authService)

	routes.RegisterRoutes(authHandler)

	port := os.Getenv("APP_PORT")
	logger.Info("Service de autenticação rodando", "port", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		logger.Info("Erro ao iniciar o serviço de authHandler.Meautenticação: ", "err", err)
	}
}

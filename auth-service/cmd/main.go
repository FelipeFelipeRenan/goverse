package main

import (
	"net/http"
	"os"

	"github.com/FelipeFelipeRenan/goverse/auth-service/internal/auth/handler"
	"github.com/FelipeFelipeRenan/goverse/auth-service/internal/auth/repository"
	"github.com/FelipeFelipeRenan/goverse/auth-service/internal/auth/service"
	"github.com/FelipeFelipeRenan/goverse/auth-service/middleware"
	"github.com/FelipeFelipeRenan/goverse/auth-service/pkg/logger"
	userpb "github.com/FelipeFelipeRenan/goverse/proto/user"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {

	logger.Init()
	godotenv.Load(".env")
	grpc_host := os.Getenv("GRPC_SERVER_HOST")
	grpc_port := os.Getenv("GRPC_SERVER_PORT")
	conn, err := grpc.NewClient(grpc_host+grpc_port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error.Fatalf("falha ao conectar ao user-service: %v", err)
	}
	defer conn.Close()

	userClient := userpb.NewUserServiceClient(conn)

	authRepo := repository.NewAuthRepository(userClient)

	passwordAuth := service.NewPasswordAuth(authRepo)

	authMethods := map[string]service.AuthMethod{
		"password": passwordAuth,
		"oauth":    service.NewOAuthAuth(authRepo),
	}
	jwt_secret := os.Getenv("JWT_SECRET")
	authService := service.NewAuthService(authMethods, []byte(jwt_secret))

	authHandler := handler.NewAuthHandler(authService)

	http.HandleFunc("POST /login", middleware.Logging(authHandler.Login))
	http.HandleFunc("/oauth/google/login", middleware.Logging(authHandler.GoogleLogin))
	http.HandleFunc("/oauth/google/callback", middleware.Logging(authHandler.GoogleCallback))

	port := os.Getenv("APP_PORT")
	logger.Info.Printf("Service de autenticação rodando na porta %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		logger.Error.Fatalf("Erro ao iniciar o serviço de autenticação: %v", err)
	}
}

package main

import (
	"log"
	"net/http"
	"os"

	"github.com/FelipeFelipeRenan/goverse/api-gateway/internal/delivery"
	"github.com/FelipeFelipeRenan/goverse/api-gateway/internal/proxy"
	"github.com/FelipeFelipeRenan/goverse/api-gateway/middleware"
	"github.com/FelipeFelipeRenan/goverse/api-gateway/pkg/logger"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env")

	logger.Init()
	mux := http.NewServeMux()

	// Rotas do auth service
	mux.Handle("/oauth/google/callback", middleware.LoggingMiddleware(proxy.NewReverseProxy("http://auth-service:8081")))


	mux.Handle("/", middleware.LoggingMiddleware(http.HandlerFunc(delivery.RouteRequest)))

	port := os.Getenv("GATEWAY_PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info.Info("API Gateway rodando", "port", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Erro ao iniciar API Gateway: %v", err)
	}
}

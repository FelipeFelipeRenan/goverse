package main

import (
	"log"
	"net/http"
	"os"

	"github.com/FelipeFelipeRenan/goverse/api-gateway/internal/proxy"
	"github.com/FelipeFelipeRenan/goverse/api-gateway/middleware"
	"github.com/FelipeFelipeRenan/goverse/api-gateway/pkg/logger"
)


func main() {

	logger.Init()
	mux := http.NewServeMux()

	mux.Handle("/auth", middleware.LoggingMiddleware(proxy.NewReverseProxy("http://auth-service:8081")))

	mux.Handle("/users", middleware.LoggingMiddleware(proxy.NewReverseProxy("http://user-service:8080")))

	port := os.Getenv("GATEWAY_PORT")
	if port == ""{
		port = "8080"
	}

	logger.Info.Info("API Gateway rodando", "port", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil{
		log.Fatalf("Erro ao iniciar API Gateway: %v", err)
	}
}
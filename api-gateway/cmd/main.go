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
	//mux.Handle("/", middleware.LoggingMiddleware(http.HandlerFunc(delivery.RouteRequest)))

	mux.Handle("/oauth/google/callback", middleware.LoggingMiddleware(proxy.NewReverseProxy("http://auth-service:8081")))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		publicPaths := []string{
			"/login",
			"/oauth/google/login",
			"/users",
			"/user",
		}

		for _, p := range publicPaths {
			if r.Method == http.MethodPost && p == "/login" && path == p {
				delivery.RouteRequest(w, r)
				return
			}
			if r.Method == http.MethodGet && p == "/oauth/google/login" && path == p {
				delivery.RouteRequest(w, r)
				return
			}
			if r.Method == http.MethodGet && p == "/users" && path == p {
				delivery.RouteRequest(w, r)
				return
			}
			if r.Method == http.MethodPost && p == "/user" && path == p {
				delivery.RouteRequest(w, r)
				return
			}
		}
		authenticaded := middleware.AuthMiddleware(http.HandlerFunc(delivery.RouteRequest))
		authenticaded.ServeHTTP(w, r)
	})

	mux.Handle("/", middleware.LoggingMiddleware(handler))

	port := os.Getenv("GATEWAY_PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info.Info("API Gateway rodando", "port", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Erro ao iniciar API Gateway: %v", err)
	}
}

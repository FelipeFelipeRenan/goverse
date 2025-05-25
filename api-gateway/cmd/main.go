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

	httpSwagger "github.com/swaggo/http-swagger"
    _ "github.com/FelipeFelipeRenan/goverse/api-gateway/docs"
)


// ref: https://swaggo.github.io/swaggo.io/declarative_comments_format/general_api_info.html
// @title Goverse API GAteway
// @version 1.0
// @description Documentação unificada dos serviços do Goverse
// @termsOfService http://swagger.io/terms/

// @contact.name Felipe Renan
// @contact.email feliperenanqwerty@gmail.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8088
// @BasePath /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	godotenv.Load(".env")

	logger.Init()
	mux := http.NewServeMux()

	// Rotas do auth service
	mux.Handle("/oauth/google/callback", middleware.LoggingMiddleware(proxy.NewReverseProxy("http://auth-service:8081")))

	mux.Handle("/swagger/", httpSwagger.WrapHandler)

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

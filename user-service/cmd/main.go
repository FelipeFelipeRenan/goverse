package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/delivery/rest/routes"
	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/handler"
	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/repository"
	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/service"
	"github.com/FelipeFelipeRenan/goverse/user-service/pkg/database"
)

func main() {
	conn, err := database.Connect()
	if err != nil {
		log.Fatalf("Erro ao conectar com banco de dados: %v", err)
	}
	defer conn.Close(nil)

	database.RunMigration(conn)

	repo := repository.NewUserRepository(conn)
	userService := service.NewUserService(repo)
	userHandler := handler.NewUserHandler(userService)

	routes.SetupUserRoutes(userHandler)

	// Iniciando o servidor na porta 8080
	port := os.Getenv("APP_PORT")
	fmt.Printf("Servidor rodando na porta %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}

}

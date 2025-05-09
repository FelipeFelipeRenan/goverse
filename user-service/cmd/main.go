package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/domain"
	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/repository"
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
	testUser := domain.User{
		Username: "Joao Silva",
		Email:    "joao@exampleeeeee.com",
		Password: "senha123",
	}

	createdID, err := repo.CreateUser(context.Background(), testUser)
	if err != nil {
		log.Fatalf("Erro ao criar usuario: %v", err)
	}
	log.Printf("Usuario criado com ID %s", createdID)

	user, err := repo.GetUserByID(context.Background(), createdID)
	if err != nil {
		log.Fatalf("Erro ao buscar usuario: %v", err)
	}
	log.Printf("Usuario buscado: %+v", user)

	// Iniciando o servidor na porta 8080
	port := os.Getenv("APP_PORT")
	fmt.Printf("Servidor rodando na porta %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}

}

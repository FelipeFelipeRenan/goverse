package main

import (
	"log"

	"github.com/FelipeFelipeRenan/goverse/user-service/pkg/database"
)

func main() {

	conn, err := database.Connect()
	if err != nil {
		log.Fatalf("Erro ao conectar com banco de dados: %v", err)
	}
	defer conn.Close(nil)

	
	
}
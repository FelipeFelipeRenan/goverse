package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/FelipeFelipeRenan/goverse/user-service/pkg/database"
)

func main() {

	conn, err := database.Connect()
	if err != nil {
		log.Fatalf("Erro ao conectar com banco de dados: %v", err)
	}
	defer conn.Close(nil)

		// Definindo uma rota de exemplo
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// Verificando a conexão com o banco
			err := conn.Ping(nil)
			if err != nil {
				http.Error(w, "Erro ao conectar ao banco de dados", http.StatusInternalServerError)
				return
			}
			w.Write([]byte("Conexão bem-sucedida com o banco de dados!"))
		})
	
		// Iniciando o servidor na porta 8080
		port := ":8080"
		fmt.Printf("Servidor rodando na porta %s...\n", port)
		if err := http.ListenAndServe(port, nil); err != nil {
			log.Fatalf("Erro ao iniciar o servidor: %v", err)
		}


	
}
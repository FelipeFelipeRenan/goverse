package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func LoadEnv() error{
	err := godotenv.Load(".env")
	if err != nil {
		return fmt.Errorf("Erro ao carregar configurações do banco de dados: %v", err)
	}
	return nil
}


func Connect() (*pgx.Conn, error){

	err := LoadEnv()
	if err != nil {
		return nil, err
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUser, dbPassword, dbHost, dbPort, dbName)
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("Não foi possível conectar ao banco de dados: %v", err)
	}

	return conn, nil
}
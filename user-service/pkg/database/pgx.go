package database

import (
	"context"
	"fmt"
	"os"

	"github.com/FelipeFelipeRenan/goverse/user-service/pkg/logger"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func LoadEnv() error {
	err := godotenv.Load(".env")
	if err != nil {
		return fmt.Errorf("Erro ao carregar configurações do banco de dados: %v", err)
	}
	return nil
}

func Connect() (*pgx.Conn, error) {

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

func RunMigration(conn *pgx.Conn) error {
	ctx := context.Background()

	query := `
		CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		picture TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		is_oauth BOOLEAN
	);	
	`

	_, err := conn.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("erro ao executar migração: %w", err)
	}

	logger.Info.Info("Migração executada com sucesso...")
	return nil
}

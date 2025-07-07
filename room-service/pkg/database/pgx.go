package database

import (
	"context"
	"fmt"
	"os"

	"github.com/FelipeFelipeRenan/goverse/room-service/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func LoadEnv() error {
	err := godotenv.Load(".env")
	if err != nil {
		return fmt.Errorf("Erro ao carregar configurações do banco de dados: %v", err)
	}
	return nil
}

func Connect() (*pgxpool.Pool, error) {
	if err := LoadEnv(); err != nil {
		return nil, err
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)

	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("não foi possível criar o pool de conexões: %v", err)
	}

	return pool, nil
}

func RunMigration(pool *pgxpool.Pool) error {
	ctx := context.Background()

	query := `
	CREATE TABLE IF NOT EXISTS rooms (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		is_public BOOLEAN NOT NULL DEFAULT true,
		max_members INT NOT NULL CHECK (max_members > 0),
		owner_id TEXT NOT NULL,
		member_count INT NOT NULL DEFAULT 0,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP NULL
	);

	CREATE TABLE IF NOT EXISTS room_members (
		room_id INT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
		user_id TEXT NOT NULL,
		role TEXT NOT NULL,
		joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (room_id, user_id)
	);
	`

	_, err := pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("erro ao executar migração: %w", err)
	}

	logger.Info("Migração executada com sucesso.")
	return nil
}

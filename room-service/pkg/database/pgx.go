package database

import (
	"context"
	"fmt"
	"os"
	"time"

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
	// if os.Getenv("ENV") != "prod" {
	// 	if err := godotenv.Load(); err != nil {
	// 		logger.Error("Erro ao carregar .env", "err", err)
	// 	}
	// }

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)

	var pool *pgxpool.Pool
	var err error

	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		pool, err = pgxpool.New(context.Background(), connStr)
		if err == nil {
			// Tenta fazer um ping para garantir que a conexão está realmente viva
			if pingErr := pool.Ping(context.Background()); pingErr == nil {
				logger.Info("Conexão com o banco de dados estabelecida com sucesso")
				return pool, nil
			} else {
				err = pingErr // Guarda o erro do ping para o log
			}
		}

		logger.Warn("Não foi possível conectar ao banco de dados. Tentando novamente...",
			"tentativa", i+1,
			"max_tentativas", maxRetries,
			"erro", err.Error(),
		)
		time.Sleep(5 * time.Second) // Espera 5 segundos antes de tentar de novo
	}

	return nil, fmt.Errorf("não foi possível conectar ao banco de dados após %d tentativas: %w", maxRetries, err)
}

func RunMigration(pool *pgxpool.Pool) error {
	ctx := context.Background()

	_, err := pool.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)
	if err != nil {
		return fmt.Errorf("erro ao ativar extensão uuid-ossp: %w", err)
	}

	query := `
	CREATE TABLE IF NOT EXISTS rooms (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		name TEXT NOT NULL,
		description TEXT,
		is_public BOOLEAN NOT NULL DEFAULT true,
		max_members INT NOT NULL CHECK (max_members > 0),
		owner_id UUID NOT NULL,
		member_count INT NOT NULL DEFAULT 0,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP NULL
	);

	CREATE TABLE IF NOT EXISTS room_members (
		room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
		user_id UUID NOT NULL,
		role TEXT NOT NULL,
		joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (room_id, user_id)
	);
	`

	_, err = pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("erro ao executar migração: %w", err)
	}

	logger.Info("Migração executada com sucesso.")
	return nil
}

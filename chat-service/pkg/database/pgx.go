package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/FelipeFelipeRenan/goverse/chat-service/pkg/logger" // Usa o novo logger
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// Connect estabelece a conexão com o banco de dados com lógica de retry.
func Connect() (*pgxpool.Pool, error) {
	// Carrega .env apenas em ambiente de desenvolvimento
	if os.Getenv("ENV") != "prod" {
		if err := godotenv.Load(); err != nil {
			logger.Warn("Arquivo .env não encontrado, usando variáveis de ambiente do sistema.")
		}
	}

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
			if pingErr := pool.Ping(context.Background()); pingErr == nil {
				logger.Info("Conexão com o banco de dados estabelecida com sucesso")
				return pool, nil
			} else {
				err = pingErr
			}
		}

		logger.Warn("Não foi possível conectar ao banco de dados. Tentando novamente...",
			"tentativa", i+1,
			"max_tentativas", maxRetries,
			"erro", err.Error(),
		)
		time.Sleep(5 * time.Second)
	}

	return nil, fmt.Errorf("não foi possível conectar ao banco de dados após %d tentativas: %w", maxRetries, err)
}

func RunMigration(pool *pgxpool.Pool) error {
	ctx := context.Background()

	// --- Comando 1: Criar a Tabela ---
	createTableQuery := `
        CREATE TABLE IF NOT EXISTS messages (
            id SERIAL PRIMARY KEY,
            room_id INT NOT NULL,
            user_id INT NOT NULL,
            username VARCHAR(255) NOT NULL,
            content TEXT NOT NULL,
            created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
        );
    `
	_, err := pool.Exec(ctx, createTableQuery)
	if err != nil {
		return fmt.Errorf("erro ao executar migração da tabela messages: %w", err)
	}

	// --- Comando 2: Criar o Índice ---
	createIndexQuery := `
        CREATE INDEX IF NOT EXISTS idx_messages_room_id_created_at ON messages (room_id, created_at DESC);
    `
	_, err = pool.Exec(ctx, createIndexQuery)
	if err != nil {
		return fmt.Errorf("erro ao criar índice para a tabela messages: %w", err)
	}

	logger.Info("Migração da tabela 'messages' executada com sucesso.")
	return nil
}

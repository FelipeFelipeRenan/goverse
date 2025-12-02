package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/FelipeFelipeRenan/goverse/common/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect estabelece conexão com retry
func Connect() (*pgxpool.Pool, error) {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

    if dbHost == "" || dbUser == "" {
        return nil, fmt.Errorf("configurações de banco de dados insuficientes")
    }

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

		logger.Warn("Falha ao conectar ao banco. Tentando novamente...",
			"tentativa", i+1,
			"erro", err.Error(),
		)
		time.Sleep(3 * time.Second)
	}

	return nil, fmt.Errorf("falha final ao conectar ao banco: %w", err)
}
package service

import (
	"context"

	"github.com/FelipeFelipeRenan/goverse/room-service/internal/repository"
	"github.com/jackc/pgx/v5"
)

// DBPool define a interface que nosso serviço precisa do pool de conexões.
type DBPool interface {
	repository.DBTX // "Herda" os métodos Exec, Query, QueryRow
	Begin(ctx context.Context) (pgx.Tx, error)
}

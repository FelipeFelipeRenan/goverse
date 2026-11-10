// Package module define o contrato que cada módulo de domínio (identity, rooms, chat)
// implementa para se plugar no processo. Módulos NÃO importam uns aos outros por pacotes
// internos: conversam apenas por interfaces pequenas exportadas na raiz de cada módulo.
package module

import (
	"context"
	"net/http"
)

type Module interface {
	Name() string
	// Register adiciona as rotas HTTP do módulo ao mux compartilhado.
	Register(mux *http.ServeMux)
}

// Migrator é implementado por módulos que possuem schema próprio.
type Migrator interface {
	Migrate(ctx context.Context) error
}

// Runner é implementado por módulos com trabalho em segundo plano (consumidores, filas, reprocessamento).
// Run bloqueia até ctx ser cancelado.
type Runner interface {
	Run(ctx context.Context) error
}

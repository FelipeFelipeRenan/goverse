package middleware

import (
	"net/http"

	"github.com/FelipeFelipeRenan/goverse/api-gateway/pkg/logger"
)

func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error.Info("🔥 Panic recuperado", "err", err)
				http.Error(w, "Erro interno no servidor", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}


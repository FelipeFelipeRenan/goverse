package middleware

import (
	"net/http"
	"time"

	"github.com/FelipeFelipeRenan/goverse/user-service/pkg/logger"
)

func Logging(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		logger.Info.Printf("[%s] %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
		logger.Info.Printf("[%s] Finalizado em %v", r.URL.Path, time.Since(start))
	})
}

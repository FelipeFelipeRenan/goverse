package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/FelipeFelipeRenan/goverse/user-service/pkg/logger"
	"github.com/FelipeFelipeRenan/goverse/user-service/pkg/metrics"
)

// wrapMux intercepta respostas com status 404 e aplica logging/metrics
func FallbackMiddleware(service string, mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// cria um response recorder para capturar a resposta original
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		start := time.Now()
		mux.ServeHTTP(rec, r)
		duration := time.Since(start)

		// rota não encontrada
		if rec.status == http.StatusNotFound {
			logger.Info("Rota não encontrada",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.status,
			)

			metrics.HTTPRequestCount.WithLabelValues(
				service, r.Method, r.URL.Path, strconv.Itoa(rec.status),
			).Inc()

			metrics.HTTPRequestDuration.WithLabelValues(
				service, r.Method, r.URL.Path,
			).Observe(duration.Seconds())
		}
	})
}

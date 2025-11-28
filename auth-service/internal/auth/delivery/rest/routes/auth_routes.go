package routes

import (
	"net/http"
	"strconv"

	"github.com/FelipeFelipeRenan/goverse/auth-service/internal/auth/handler"
	"github.com/FelipeFelipeRenan/goverse/auth-service/middleware"
	"github.com/FelipeFelipeRenan/goverse/auth-service/pkg/metrics"
	"github.com/FelipeFelipeRenan/goverse/common/pkg/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RegisterRoutes(authHandler *handler.AuthHandler) {
	// Auth endpoints
	http.HandleFunc("POST /auth/login", withCommonMiddleware("auth-service", authHandler.Login))
	http.HandleFunc("/oauth/google/login", withCommonMiddleware("auth-service", authHandler.GoogleLogin))
	http.HandleFunc("/oauth/google/callback", withCommonMiddleware("auth-service", authHandler.GoogleCallback))

	http.HandleFunc("POST /auth/logout", withCommonMiddleware("auth-service", authHandler.Logout))

	http.HandleFunc("GET /auth/me", withCommonMiddleware("auth-service", authHandler.Me))

	// Metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	// Default 404 handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		status := http.StatusNotFound
		http.Error(w, "Rota não encontrada", status)

		logger.Info("Rota inválida",
			"method", r.Method,
			"path", r.URL.Path,
			"status", status,
		)

		metrics.HTTPRequestCount.WithLabelValues(
			"auth-service", r.Method, r.URL.Path, strconv.Itoa(status),
		).Inc()
	})
}

func withCommonMiddleware(service string, h http.HandlerFunc) http.HandlerFunc {
	return middleware.ChainMiddleware(
		middleware.MetricsMiddleware(service),
		middleware.LoggingMiddleware,
	)(h)
}

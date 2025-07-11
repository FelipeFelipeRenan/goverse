package routes

import (
	"net/http"
	"strconv"

	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/handler"
	"github.com/FelipeFelipeRenan/goverse/user-service/middleware"
	"github.com/FelipeFelipeRenan/goverse/user-service/pkg/logger"
	"github.com/FelipeFelipeRenan/goverse/user-service/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func SetupUserRoutes(userHandler *handler.UserHandler) {
	http.HandleFunc("GET /users", withCommonMiddleware("user-service", userHandler.GetAllUsers))
	http.HandleFunc("GET /user/{id}", withCommonMiddleware("user-service", userHandler.GetByID))
	http.HandleFunc("POST /user", withCommonMiddleware("user-service", userHandler.Register))
	http.HandleFunc("PUT /user/me", withCommonMiddleware("user-service", userHandler.UpdateUser))
	http.HandleFunc("DELETE /user/me", withCommonMiddleware("user-service", userHandler.DeleteUser))

	// Exposição de métricas Prometheus
	http.Handle("/metrics", promhttp.Handler())

	// Adicione ao final da SetupUserRoutes:
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		status := http.StatusNotFound
		http.Error(w, "Rota não encontrada", status)

		logger.Info("Rota inválida",
			"method", r.Method,
			"path", r.URL.Path,
			"status", status,
		)

		metrics.HTTPRequestCount.WithLabelValues(
			"user-service", r.Method, r.URL.Path, strconv.Itoa(status),
		).Inc()
	})

}

func withCommonMiddleware(service string, h http.HandlerFunc) http.HandlerFunc {
	return middleware.ChainMiddleware(
		middleware.MetricsMiddleware(service),
		middleware.LoggingMiddleware,
	)(h)
}

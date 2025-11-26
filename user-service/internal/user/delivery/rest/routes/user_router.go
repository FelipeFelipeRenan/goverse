package routes

import (
	"net/http"

	"github.com/FelipeFelipeRenan/goverse/common/pkg/logger"
	"github.com/FelipeFelipeRenan/goverse/common/pkg/middleware"
	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/handler"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func SetupUserRoutes(userHandler *handler.UserHandler) {
	// Rotas de Negócio (protegidas por middlewares)
	http.HandleFunc("GET /users", withCommonMiddleware("user-service", userHandler.GetAllUsers))
	http.HandleFunc("GET /user/{id}", withCommonMiddleware("user-service", userHandler.GetByID))
	http.HandleFunc("POST /user", withCommonMiddleware("user-service", userHandler.Register))
	http.HandleFunc("PUT /user/me", withCommonMiddleware("user-service", userHandler.UpdateUser))
	http.HandleFunc("DELETE /user/me", withCommonMiddleware("user-service", userHandler.DeleteUser))

	// Exposição de métricas Prometheus (sem middleware de log/metrics para não poluir)
	http.Handle("/metrics", promhttp.Handler())

	// Handler 404 (Fallback)
	// Usamos o middleware aqui também para garantir que erros 404 sejam contabilizados nas métricas
	http.HandleFunc("/", withCommonMiddleware("user-service", func(w http.ResponseWriter, r *http.Request) {
		status := http.StatusNotFound
		http.Error(w, "Rota não encontrada", status)

		// Log específico de aviso (o middleware já fará o log estruturado da requisição)
		logger.Info("Rota inexistente acessada",
			"method", r.Method,
			"path", r.URL.Path,
			"status", status,
		)
	}))
}

// withCommonMiddleware aplica a cadeia de middlewares padrão do Common
func withCommonMiddleware(service string, h http.HandlerFunc) http.HandlerFunc {
	return middleware.Chain(
		middleware.Metrics(service), // Define o nome do serviço para as labels do Prometheus
		middleware.Logging,          // Loga a requisição e o tempo de resposta
	)(h)
}
package routes

import (
	"net/http"

	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/handler"
	"github.com/FelipeFelipeRenan/goverse/user-service/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func SetupUserRoutes(userHandler *handler.UserHandler) {
	http.HandleFunc("GET /users", withCommonMiddleware("user-service", userHandler.GetAllUsers))
	http.HandleFunc("GET /user/{id}", withCommonMiddleware("user-service", userHandler.GetByID))
	http.HandleFunc("POST /user", withCommonMiddleware("user-service", userHandler.Register))
	http.HandleFunc("PUT /user/me", withCommonMiddleware("user-service", userHandler.UpdateUser))
	http.HandleFunc("DELETE /user/me", withCommonMiddleware("user-service", userHandler.DeleteUser))

	// Métricas sem middleware
	http.Handle("/metrics", promhttp.Handler())
}



// routes/utils.go ou dentro do SetupUserRoutes mesmo
func withCommonMiddleware(service string, h http.HandlerFunc) http.HandlerFunc {
	return middleware.ChainMiddleware(
		middleware.MetricsMiddleware(service),
		middleware.LoggingMiddleware,
	)(h)
}

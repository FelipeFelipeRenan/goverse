package routes

import (
	"net/http"

	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/handler"
	"github.com/FelipeFelipeRenan/goverse/user-service/middleware"
)

func SetupUserRoutes(userHandler *handler.UserHandler) {

	http.HandleFunc("GET /users", middleware.LoggingMiddleware(userHandler.GetAllUsers))
	http.HandleFunc("GET /user/{id}", middleware.LoggingMiddleware(userHandler.GetByID))
	http.HandleFunc("POST /user", middleware.LoggingMiddleware(userHandler.Register))
	http.HandleFunc("PUT /user/{id}", middleware.LoggingMiddleware(userHandler.UpdateUser))
	http.HandleFunc("DELETE /user/{id}", middleware.LoggingMiddleware(userHandler.DeleteUser))
}

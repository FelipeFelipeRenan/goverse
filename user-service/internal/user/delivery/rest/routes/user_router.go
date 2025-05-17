package routes

import (
	"net/http"

	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/handler"
	"github.com/FelipeFelipeRenan/goverse/user-service/middleware"
)

func SetupUserRoutes(userHandler *handler.UserHandler) {

	http.HandleFunc("GET /users", middleware.Logging(userHandler.GetAllUsers))
	http.HandleFunc("GET /user/{id}", middleware.Logging(userHandler.GetByID))
	http.HandleFunc("POST /user", middleware.Logging(userHandler.Register))
}

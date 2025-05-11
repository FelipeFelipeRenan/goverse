package routes

import (
	"net/http"

	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/handler"
)

func SetupUserRoutes(userHandler *handler.UserHandler) {

	http.HandleFunc("GET /users", userHandler.GetAllUsers)
	http.HandleFunc("GET /user/{id}", userHandler.GetByID)
	http.HandleFunc("POST /user", userHandler.Register)
}

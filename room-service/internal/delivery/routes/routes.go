package routes

import (
	"net/http"

	"github.com/FelipeFelipeRenan/goverse/room-service/internal/handler"
	"github.com/FelipeFelipeRenan/goverse/room-service/middleware"
)

func RegisterRoutes(roomHandler *handler.RoomHandler) {

	http.HandleFunc("POST /rooms", middleware.LoggingMiddleware(roomHandler.CreateRoom))
	http.HandleFunc("GET /rooms/{id}", middleware.LoggingMiddleware(roomHandler.GetRoomByID))
	http.HandleFunc("GET /rooms", middleware.LoggingMiddleware(roomHandler.ListPublicRooms))

}

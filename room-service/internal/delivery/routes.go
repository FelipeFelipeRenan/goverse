package delivery

import (
	"net/http"

	"github.com/FelipeFelipeRenan/goverse/room-service/internal/handler"
)

func RegisterRoutes(roomHandler *handler.RoomHandler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /rooms", roomHandler.CreateRoom)
	mux.HandleFunc("GET /rooms/{id}", roomHandler.GetRoomByID)

	return mux
}

package routes

import (
	"net/http"

	"github.com/FelipeFelipeRenan/goverse/room-service/internal/handler"
	"github.com/FelipeFelipeRenan/goverse/room-service/middleware"
)

func RegisterRoutes(roomHandler *handler.RoomHandler, memberHandler *handler.MemberHandler) {

	// Rotas de rooms
	http.HandleFunc("POST /rooms", middleware.LoggingMiddleware(roomHandler.CreateRoom))
	http.HandleFunc("GET /rooms/{id}", middleware.LoggingMiddleware(roomHandler.GetRoomByID))
	http.HandleFunc("GET /rooms", middleware.LoggingMiddleware(roomHandler.ListRooms))
	http.HandleFunc("PATCH /rooms/{id}", middleware.LoggingMiddleware(roomHandler.UpdateRoom))
	http.HandleFunc("DELETE /rooms/{id}", middleware.LoggingMiddleware(roomHandler.DeleteRoom))

	// Rotas de membros
	http.HandleFunc("GET /rooms/{roomID}/members", middleware.LoggingMiddleware(memberHandler.ListMembers))
	http.HandleFunc("POST /rooms/{roomID}/members", middleware.LoggingMiddleware(memberHandler.AddMember))
	http.HandleFunc("PUT /rooms/{roomID}/members/{memberID}/role", middleware.LoggingMiddleware(memberHandler.UpdateRole))
	http.HandleFunc("DELETE /rooms/{roomID}/members/{memberID}", middleware.LoggingMiddleware(memberHandler.RemoveMember))

}

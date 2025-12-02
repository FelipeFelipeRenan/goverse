package routes

import (
	"net/http"
	"strconv"

	"github.com/FelipeFelipeRenan/goverse/common/pkg/logger"
	"github.com/FelipeFelipeRenan/goverse/room-service/internal/handler"
	"github.com/FelipeFelipeRenan/goverse/room-service/middleware"
	"github.com/FelipeFelipeRenan/goverse/room-service/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RegisterRoutes(roomHandler *handler.RoomHandler, memberHandler *handler.MemberHandler) {
	// Rooms
	http.HandleFunc("POST /rooms", withCommonMiddleware("room-service", roomHandler.CreateRoom))
	http.HandleFunc("GET /rooms/{id}", withCommonMiddleware("room-service", roomHandler.GetRoomByID))
	http.HandleFunc("GET /rooms", withCommonMiddleware("room-service", roomHandler.ListRooms))
	http.HandleFunc("PATCH /rooms/{id}", withCommonMiddleware("room-service", roomHandler.UpdateRoom))
	http.HandleFunc("DELETE /rooms/{id}", withCommonMiddleware("room-service", roomHandler.DeleteRoom))

	// Members
	http.HandleFunc("POST /rooms/{roomID}/join", withCommonMiddleware("room-service", memberHandler.JoinRoom))
	http.HandleFunc("GET /rooms/{roomID}/members", withCommonMiddleware("room-service", memberHandler.ListMembers))
	http.HandleFunc("GET /user/rooms", withCommonMiddleware("room-service", memberHandler.GetRoomsByUserID))
	http.HandleFunc("GET /rooms/mine", withCommonMiddleware("room-service", memberHandler.GetRoomsOwnedByUser))
	http.HandleFunc("POST /rooms/{roomID}/members", withCommonMiddleware("room-service", memberHandler.AddMember))
	http.HandleFunc("PUT /rooms/{roomID}/members/{memberID}/role", withCommonMiddleware("room-service", memberHandler.UpdateRole))
	http.HandleFunc("DELETE /rooms/{roomID}/members/{memberID}", withCommonMiddleware("room-service", memberHandler.RemoveMember))

	http.Handle("/metrics", promhttp.Handler())

	// 404
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		status := http.StatusNotFound
		http.Error(w, "Rota não encontrada", status)

		logger.Info("Rota inválida",
			"method", r.Method,
			"path", r.URL.Path,
			"status", status,
		)

		metrics.HTTPRequestCount.WithLabelValues(
			"room-service", r.Method, r.URL.Path, strconv.Itoa(status),
		).Inc()
	})
}

func withCommonMiddleware(service string, h http.HandlerFunc) http.HandlerFunc {
	return middleware.ChainMiddleware(
		middleware.MetricsMiddleware(service),
		middleware.LoggingMiddleware,
	)(h)
}

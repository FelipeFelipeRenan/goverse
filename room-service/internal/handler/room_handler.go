package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/FelipeFelipeRenan/goverse/room-service/internal/domain"
	"github.com/FelipeFelipeRenan/goverse/room-service/internal/dtos"
	"github.com/FelipeFelipeRenan/goverse/room-service/internal/service"
	"github.com/google/uuid"
)

type RoomHandler struct {
	RoomService service.RoomService
}

func NewRoomHandler(roomService service.RoomService) *RoomHandler {
	return &RoomHandler{
		RoomService: roomService,
	}
}

func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req dtos.CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("falha ao ler corpo da requisição: %v", err))
		return
	}

	ownerID, err := uuid.Parse(req.OwnerID)
	if err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("ID do dono inválido: %v", err))
		return
	}

	room := &domain.Room{
		Name:        req.Name,
		Description: req.Description,
		IsPublic:    req.IsPublic,
		MaxMembers:  req.MaxMembers,
	}

	createdRoom, err := h.RoomService.CreateRoom(r.Context(), ownerID.String(), room)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := dtos.FromRoom(createdRoom)

	sendResponse(w, http.StatusCreated, resp)

}

func (h *RoomHandler) GetRoomByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	roomID, err := uuid.Parse(idStr)
	if err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("ID de sala inválido: %v", err))
		return
	}

	room, err := h.RoomService.GetRoomByID(r.Context(), roomID.String())
	if err != nil {
		sendError(w, http.StatusNotFound, err.Error())
		return
	}

	resp := dtos.FromRoom(room)

	sendResponse(w, http.StatusOK, resp)
}

func sendError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func sendResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

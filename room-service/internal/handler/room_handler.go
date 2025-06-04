package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/FelipeFelipeRenan/goverse/room-service/internal/domain"
	"github.com/FelipeFelipeRenan/goverse/room-service/internal/dtos"
	"github.com/FelipeFelipeRenan/goverse/room-service/internal/service"
)

type RoomHandler struct {
	RoomService   service.RoomService
	UserValidator service.UserValidator
}

func NewRoomHandler(roomService service.RoomService, validator service.UserValidator) *RoomHandler {
	return &RoomHandler{
		RoomService:   roomService,
		UserValidator: validator,
	}
}

func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req dtos.CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("falha ao ler corpo da requisição: %v", err))
		return
	}

	ownerID := req.OwnerID
	if ownerID == "" {
		sendError(w, http.StatusBadRequest, "ID do dono inválido")
		return
	}
	valid, err := h.UserValidator.IsUserValid(r.Context(), ownerID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("erro ao validar o usuário: %v", err))
		return
	}
	if !valid {
		sendError(w, http.StatusBadRequest, "owner_id inválido")
		return
	}

	room := &domain.Room{
		Name:        req.Name,
		Description: req.Description,
		IsPublic:    req.IsPublic,
		MaxMembers:  req.MaxMembers,
	}
	if room.MaxMembers <= 0 {
		room.MaxMembers = 10 // valor padrao para nao quebrar constraint da migração max_members > 0
	}

	room.MemberCount = 1
	createdRoom, err := h.RoomService.CreateRoom(r.Context(), ownerID, room)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := dtos.FromRoom(createdRoom)

	sendResponse(w, http.StatusCreated, resp)

}

func (h *RoomHandler) DeleteRoom(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	roomID := r.PathValue("id")
	if roomID == "" {
		sendError(w, http.StatusBadRequest, "room_id é obrigatório")
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		sendError(w, http.StatusUnauthorized, "Falta user ID")
		return
	}

	err := h.RoomService.DeleteRoom(ctx, userID, roomID)
	if err != nil {
		if errors.Is(err, domain.ErrRoomNotFound) {
			sendError(w, http.StatusNotFound, "Sala não encontrada")
			return
		}
		if strings.Contains(err.Error(), "somente o dono") {
			sendError(w, http.StatusForbidden, err.Error())
			return
		}
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("Erro ao deletar sala: %v", err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *RoomHandler) GetRoomByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	roomID := idStr
	if roomID == "" {
		sendError(w, http.StatusBadRequest, "ID de sala inválido")
		return
	}

	room, err := h.RoomService.GetRoomByID(r.Context(), roomID)
	if err != nil {
		sendError(w, http.StatusNotFound, err.Error())
		return
	}

	if room == nil {
		sendError(w, http.StatusNotFound, "room not found")
		return
	}

	resp := dtos.FromRoom(room)

	sendResponse(w, http.StatusOK, resp)
}

func (h *RoomHandler) ListRooms(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit := 10 // valor default
	offset := 0
	publicOnly := true
	keyword := ""

	// parse do limite
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}

	// parse do offset
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	// parse do publicOnly
	if p := r.URL.Query().Get("public_only"); p != "" {
		publicOnly = (p == "true" || p == "1")
	}

	// parse do keyword
	keyword = r.URL.Query().Get("keyword")

	rooms, err := h.RoomService.ListRooms(ctx, limit, offset, publicOnly, keyword)
	if err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("erro ao listar salas: %v", err))
		return
	}

	sendResponse(w, http.StatusOK, rooms)

}

func (h *RoomHandler) UpdateRoom(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	roomID := r.PathValue("id")
	if roomID == "" {
		sendError(w, http.StatusBadRequest, "room_id é obrigatorio")
		return
	}

	// Extrair userID do contexto ou header
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		sendError(w, http.StatusUnauthorized, "Falta user ID")
		return
	}

	var req struct {
		Name        *string `json:"name,omitempty"`
		Description *string `json:"description,omitempty"`
		IsPublic    *bool   `json:"is_public,omitempty"`
		MaxMembers  *int    `json:"max_members,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Corpo de requisição inválido")
		return
	}

	existingRoom, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		if errors.Is(err, domain.ErrRoomNotFound) {
			sendError(w, http.StatusNotFound, "Sala não encontrada")
			return
		}
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("Erro ao buscar sala: %v", err))
		return
	}

	// Merge dos campos atualizaveis, apenas os fornecidos
	if req.Name != nil {
		existingRoom.Name = *req.Name
	}
	if req.Description != nil {
		existingRoom.Description = *req.Description
	}
	if req.IsPublic != nil {
		existingRoom.IsPublic = *req.IsPublic
	}
	if req.MaxMembers != nil {
		existingRoom.MaxMembers = *req.MaxMembers
	}

	if err := h.RoomService.UpdateRoom(ctx, userID, existingRoom); err != nil {
		if strings.Contains(err.Error(), "não tem permissão") {
			sendError(w, http.StatusForbidden, "Sem permissão para editar esta sala")
			return
		}
		sendError(w, http.StatusInternalServerError, "Erro ao atualizar sala")
		return
	}

	sendResponse(w, http.StatusOK, existingRoom)
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

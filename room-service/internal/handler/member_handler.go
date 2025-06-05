package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/FelipeFelipeRenan/goverse/room-service/internal/dtos"
	"github.com/FelipeFelipeRenan/goverse/room-service/internal/service"
)

type MemberHandler struct {
	memberService service.MemberService
}

func NewMemberHandler(memberService service.MemberService) *MemberHandler {
	return &MemberHandler{
		memberService: memberService,
	}
}

func (h *MemberHandler) AddMember(w http.ResponseWriter, r *http.Request) {

	roomID := r.PathValue("roomID")
	actorID := r.Header.Get("X-User-ID")

	var req dtos.AddMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Corpo de requisição inválido")
		return
	}

	if err := h.memberService.AddMember(r.Context(), actorID, roomID, req.UserID, req.Role); err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("Erro ao adicionar membro à sala: %v", err))
		return
	}

	sendResponse(w, http.StatusCreated, req)
}

func (h *MemberHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {

	roomID := r.PathValue("roomID")
	userID := r.PathValue("memberID")
	actorID := r.Header.Get("X-User-ID")

	if err := h.memberService.RemoveMember(r.Context(), actorID, roomID, userID); err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("Erro ao remover membro: %v", err))
		return
	}

	sendResponse(w, http.StatusOK, userID)

}

func (h *MemberHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("roomID")
	userID := r.PathValue("memberID")
	actorID := r.Header.Get("X-User-ID")

	var req dtos.UpdateRoleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Corpo de requisição inválido")
		return
	}

	if err := h.memberService.UpdateMemberRole(r.Context(), actorID, roomID, userID, req.NewRole); err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("Erro ao realizar upgrade no role: %v", err))
		return
	}

	sendResponse(w, http.StatusOK, req)

}

func (h *MemberHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("roomID")

	members, err := h.memberService.GetRoomMembers(r.Context(), roomID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("Erro ao receber membros da sala: %v", err))
		return
	}

	sendResponse(w, http.StatusOK, members)
}

// func sendError(w http.ResponseWriter, statusCode int, message string) {
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(statusCode)
// 	json.NewEncoder(w).Encode(map[string]string{"error": message})
// }

// func sendResponse(w http.ResponseWriter, statusCode int, data interface{}) {
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(statusCode)

// 	if data != nil {
// 		json.NewEncoder(w).Encode(data)
// 	}
// }

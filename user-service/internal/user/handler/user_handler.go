package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/domain"
	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/service"
)

type UserHandler struct {
	Service service.UserService
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{
		Service: svc,
	}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var user domain.User

	body, err := io.ReadAll(r.Body)
	if err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("falha ao ler corpo da requisição: %v", err))
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &user); err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("formato de requisição inválido: %v", err))
		return
	}

	id, err := h.Service.Register(r.Context(), user)
	if err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("falha ao registrar usuário: %v", err))
		return
	}
	sendResponse(w, http.StatusCreated, map[string]string{"user": id.ID})
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "Usuário não autenticado", http.StatusUnauthorized)
		return
	}

	if userID == "" {
		sendError(w, http.StatusBadRequest, "falha ao solicitar usuario: id vazio")
		return
	}
	var input domain.User
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("erro ao ler corpo da requisição: %v", err))
		return
	}

	updatedUser, err := h.Service.UpdateUser(r.Context(), userID, input)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			sendError(w, http.StatusNotFound, "usuário não encontrado")
			return
		}
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("Erro ao atualizar usuário: %v", err))
		return
	}

	sendResponse(w, http.StatusOK, updatedUser)

}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "Usuário não autenticado", http.StatusUnauthorized)
		return
	}

	err := h.Service.DeleteUser(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			sendError(w, http.StatusNotFound, "usuário não encontrado")
			return
		}
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("erro ao deleter usuário: %v", err))
		return
	}
	sendResponse(w, http.StatusOK, userID)
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		sendError(w, http.StatusBadRequest, "falha ao solicitar usuario: id vazio")
		return
	}
	user, err := h.Service.FindByID(r.Context(), id)
	if err != nil {
		sendError(w, http.StatusNotFound, "usuario nao encontrado")
		return
	}
	sendResponse(w, http.StatusOK, user)
}

func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.Service.GetAllUsers(r.Context())
	if users == nil {
		users = []domain.User{}
	}

	if err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("falha ao solicitar usuarios: %v", err))
		return
	}
	sendResponse(w, http.StatusOK, users)
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

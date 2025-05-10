package handler

import (
	"encoding/json"
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
		sendError(w, http.StatusBadRequest, "falha ao ler corpo da requisição")
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &user); err != nil {
		sendError(w, http.StatusBadRequest, "formato de requisição invalida")
		return
	}

	id, err := h.Service.Register(r.Context(), user)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "falha ao registrar usuario")
	}
	sendResponse(w, http.StatusCreated, map[string]string{"id": id})
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
		sendError(w, http.StatusInternalServerError, "falha ao solicitar usuarios")
		return
	}
	sendResponse(w, http.StatusOK, users)
}

func sendError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func sendResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

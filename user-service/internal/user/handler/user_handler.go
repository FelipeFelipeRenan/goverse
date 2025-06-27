package handler

import (
	"crypto/rand"
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
	// Se a senha estiver vazia, gera uma senha aleatória para evitar erro no serviço
	if user.Password == "" {
		user.Password, err = generateRandomPassword(16)
		if err != nil {
			sendError(w, http.StatusInternalServerError, fmt.Sprintf("erro ao gerar senha automática: %v", err))
			return
		}
	}

	id, err := h.Service.Register(r.Context(), user)
	if err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("falha ao registrar usuário: %v", err))
		return
	}
	sendResponse(w, http.StatusCreated, map[string]string{"user": id.ID})
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if id == "" {
		sendError(w, http.StatusBadRequest, "falha ao solicitar usuario: id vazio")
		return
	}
	var input domain.User
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("erro ao ler corpo da requisição: %v", err))
		return
	}

	updatedUser, err := h.Service.UpdateUser(r.Context(), id, input)
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

// Essa função é uma gambiarra
func generateRandomPassword(length int) (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	for i := range bytes {
		bytes[i] = letters[bytes[i]%byte(len(letters))]
	}
	return string(bytes), nil
}

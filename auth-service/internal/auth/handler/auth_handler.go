package handler

import (
	"encoding/json"
	"net/http"

	"github.com/FelipeFelipeRenan/goverse/auth-service/internal/auth/domain"
	"github.com/FelipeFelipeRenan/goverse/auth-service/internal/auth/service"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler{
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request){
	var creds domain.Credentials

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil{
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	tokenResp, err := h.authService.Authenticate(r.Context(), creds)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokenResp)
}
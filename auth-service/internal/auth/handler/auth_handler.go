package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/FelipeFelipeRenan/goverse/auth-service/internal/auth/domain"
	"github.com/FelipeFelipeRenan/goverse/auth-service/internal/auth/service"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var creds domain.Credentials

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
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

func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	state := "state-token" // idealmente aleatório e salvo em cookie/session
	authURL := h.authService.GetOAuthURL(state)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "code ausente", http.StatusBadRequest)
		return
	}

	tokenResp, err := h.authService.Authenticate(ctx, domain.Credentials{
		Type:  "oauth",
		Token: code,
	})
	if err != nil {
		http.Error(w, "erro na autenticação: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Retorna HTML com postMessage
	html := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head><title>Autenticado</title></head>
		<body>
			<script>
				window.opener.postMessage({
					type: 'oauth-success',
					token: '%s'
				}, '*');
				window.close();
			</script>
		</body>
		</html>
	`, tokenResp.Token)

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

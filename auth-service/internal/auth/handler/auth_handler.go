package handler

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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

	tokenString, user, err := h.authService.Authenticate(r.Context(), creds)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(24 * time.Hour)

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    tokenString,
		Expires:  expirationTime,
		HttpOnly: true,
		Secure:   false, // Mude para true em produção
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	b := make([]byte, 32)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Expires:  time.Now().Add(10 * time.Minute), // Validade curta
		HttpOnly: true,
		Secure:   false, // Mude para true em produção
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	authURL := h.authService.GetOAuthURL(state)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	returnedState := r.URL.Query().Get("state")

	stateCookie, err := r.Cookie("oauth_state")
	if err != nil {
		http.Error(w, "state cookie não encontrado ou expirado", http.StatusBadRequest)
		return
	}

	if returnedState != stateCookie.Value {
		http.Error(w, "state inválido", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "code ausente", http.StatusBadRequest)
		return
	}

	tokenString, _, err := h.authService.Authenticate(ctx, domain.Credentials{
		Type:  "oauth",
		Token: code,
	})
	if err != nil {
		http.Error(w, "erro na autenticação: "+err.Error(), http.StatusUnauthorized)
		return
	}
	expirationTime := time.Now().Add(24 * time.Hour)
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    tokenString,
		Expires:  expirationTime,
		HttpOnly: true,
		Secure:   false, // Mude para true em produção
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	html := `
    <!DOCTYPE html>
    <html>
    <head><title>Autenticado</title></head>
    <body>
        <script>
            window.opener.postMessage({ type: 'oauth-success' }, '*');
            window.close();
        </script>
        <p>Login concluído! Você pode fechar esta janela.</p>
    </body>
    </html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fmt.Fprint(w, html)

}

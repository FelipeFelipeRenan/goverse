package handler

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/FelipeFelipeRenan/goverse/auth-service/internal/auth/domain"
	"github.com/FelipeFelipeRenan/goverse/auth-service/internal/auth/service"
	"github.com/google/uuid"
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

	isProd := os.Getenv("ENV") == "prod"
	sameSitePolicy := http.SameSiteLaxMode // Padrão seguro

	if isProd {
		sameSitePolicy = http.SameSiteNoneMode
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    tokenString,
		MaxAge:   86400, // 24 horas em segundos
		HttpOnly: true,
		Secure:   isProd, // true em produção, false em dev
		SameSite: sameSitePolicy,
		Path:     "/",
	})

	csrfToken := uuid.NewString()

	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    csrfToken,
		MaxAge:   86400, // Mesma duração do token de acesso
		Secure:   isProd,
		SameSite: sameSitePolicy,
		Path:     "/",
		// HttpOnly: false, // <-- IMPORTANTE: Este cookie NÃO é HttpOnly
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user":       user,
		"csrf_token": csrfToken,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	isProd := os.Getenv("ENV") == "prod"
	sameSitePolicy := http.SameSiteLaxMode

	if isProd {
		sameSitePolicy = http.SameSiteNoneMode
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		MaxAge:   -1, // Deleta o cookie agora
		HttpOnly: true,
		Secure:   isProd,
		SameSite: sameSitePolicy,
		Path:     "/",
	})

	// limpa o cookie de CSRF
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    "",
		MaxAge:   -1,
		Secure:   isProd,
		SameSite: sameSitePolicy,
		Path:     "/",
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "logout successful"})
}

func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	b := make([]byte, 32)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)

	isProd := os.Getenv("ENV") == "prod"
	sameSitePolicy := http.SameSiteLaxMode

	if isProd {
		sameSitePolicy = http.SameSiteNoneMode
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
		Secure:   isProd,
		SameSite: sameSitePolicy,
		Path:     "/",
	})

	authURL := h.authService.GetOAuthURL(state)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stateCookie, err := r.Cookie("oauth_state")
	if err != nil {
		http.Error(w, "state cookie não encontrado ou expirado", http.StatusBadRequest)
		return
	}
	if r.URL.Query().Get("state") != stateCookie.Value {
		http.Error(w, "state inválido", http.StatusBadRequest)
		return
	}

	// Limpa o cookie de state, ele só pode ser usado uma vez
	http.SetCookie(w, &http.Cookie{Name: "oauth_state", MaxAge: -1})

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

	isProd := isProd()
	sameSitePolicy := http.SameSiteLaxMode

	if isProd {
		sameSitePolicy = http.SameSiteNoneMode
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    tokenString,
		MaxAge:   86400,
		HttpOnly: true,
		Secure:   isProd,
		SameSite: sameSitePolicy,
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

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	// O user_id já foi validado e injetado pelo auth-middleware
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "Cabeçalho X-User-ID não encontrado", http.StatusUnauthorized)
		return
	}

	// Agora a chamada funciona
	user, err := h.authService.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "Usuário não encontrado", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// isProd é uma função helper para verificar a variável de ambiente
func isProd() bool {
	return os.Getenv("ENV") == "prod"
}

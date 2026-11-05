package identity

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/FelipeFelipeRenan/goverse/internal/platform/authn"
	"github.com/FelipeFelipeRenan/goverse/internal/platform/httpx"
	"golang.org/x/oauth2"
)

const stateCookie = "oauth_state"

type googleOAuth struct{ cfg *oauth2.Config }

// newGoogle devolve nil quando o Google não está configurado (as rotas respondem 501).
func newGoogle(c Config) *googleOAuth {
	if c.GoogleClientID == "" {
		return nil
	}
	return &googleOAuth{cfg: &oauth2.Config{
		ClientID:     c.GoogleClientID,
		ClientSecret: c.GoogleClientSecret,
		RedirectURL:  c.GoogleRedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}}
}

func (a *api) googleLogin(w http.ResponseWriter, r *http.Request) {
	if a.google == nil {
		httpx.Error(w, http.StatusNotImplemented, "login com Google não configurado")
		return
	}
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}
	state := base64.URLEncoding.EncodeToString(b)
	http.SetCookie(w, &http.Cookie{
		Name: stateCookie, Value: state, Path: "/", MaxAge: 600,
		HttpOnly: true, Secure: a.secure, SameSite: authn.SameSite(a.secure),
	})
	http.Redirect(w, r, a.google.cfg.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

func (a *api) googleCallback(w http.ResponseWriter, r *http.Request) {
	if a.google == nil {
		httpx.Error(w, http.StatusNotImplemented, "login com Google não configurado")
		return
	}
	c, err := r.Cookie(stateCookie)
	if err != nil || c.Value == "" || r.URL.Query().Get("state") != c.Value {
		httpx.Error(w, http.StatusBadRequest, "state inválido ou expirado")
		return
	}
	// O state só vale uma vez (mesmos atributos do cookie original para o navegador removê-lo).
	http.SetCookie(w, &http.Cookie{
		Name: stateCookie, Path: "/", MaxAge: -1,
		HttpOnly: true, Secure: a.secure, SameSite: authn.SameSite(a.secure),
	})

	code := r.URL.Query().Get("code")
	if code == "" {
		httpx.Error(w, http.StatusBadRequest, "code ausente")
		return
	}

	info, err := a.google.userInfo(r.Context(), code)
	if err != nil {
		slog.WarnContext(r.Context(), "oauth google", "err", err)
		httpx.Error(w, http.StatusUnauthorized, "falha na autenticação com Google")
		return
	}
	// Sem e-mail verificado não dá para vincular a uma conta existente com segurança.
	if !info.VerifiedEmail || info.Email == "" {
		httpx.Error(w, http.StatusUnauthorized, "e-mail do Google não verificado")
		return
	}

	email := normEmail(info.Email)
	row, err := a.st.byEmail(r.Context(), email)
	var user User
	switch {
	case err == nil:
		user = row.User
	case errors.Is(err, ErrNotFound):
		user, err = a.st.create(r.Context(), newUser{
			Username: info.Name, Email: email, Picture: info.Picture, IsOAuth: true,
		})
		if err != nil {
			slog.ErrorContext(r.Context(), "criar usuário oauth", "err", err)
			httpx.Error(w, http.StatusInternalServerError, "erro ao registrar usuário")
			return
		}
	default:
		slog.ErrorContext(r.Context(), "buscar usuário oauth", "err", err)
		httpx.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}

	// Emite também o CSRF: antes o login Google não o criava e todo POST/PUT/DELETE falhava.
	if _, ok := a.startSession(w, r, user, false); !ok {
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<!DOCTYPE html><html><head><title>Autenticado</title></head><body>
<script>window.opener && window.opener.postMessage({ type: 'oauth-success' }, '*'); window.close();</script>
<p>Login concluído! Você pode fechar esta janela.</p></body></html>`)
}

type googleUser struct {
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

func (g *googleOAuth) userInfo(ctx context.Context, code string) (googleUser, error) {
	ctx = context.WithValue(ctx, oauth2.HTTPClient, &http.Client{Timeout: 10 * time.Second})
	tok, err := g.cfg.Exchange(ctx, code)
	if err != nil {
		return googleUser{}, fmt.Errorf("trocar code por token: %w", err)
	}
	resp, err := g.cfg.Client(ctx, tok).Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return googleUser{}, fmt.Errorf("userinfo: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return googleUser{}, fmt.Errorf("userinfo: status %d", resp.StatusCode)
	}
	var u googleUser
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return googleUser{}, fmt.Errorf("decodificar userinfo: %w", err)
	}
	u.Name = strings.TrimSpace(u.Name)
	if u.Name == "" {
		u.Name = strings.SplitN(u.Email, "@", 2)[0]
	}
	return u, nil
}

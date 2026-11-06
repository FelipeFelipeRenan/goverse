package identity

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/FelipeFelipeRenan/goverse/internal/platform/authn"
	"github.com/FelipeFelipeRenan/goverse/internal/platform/httpx"
	"golang.org/x/crypto/bcrypt"
)

type api struct {
	st        store
	keys      *authn.Keys
	secure    bool
	google    *googleOAuth
	dummyHash []byte // usado para igualar o tempo de resposta quando o e-mail não existe

	// nil = sem limite (testes). Em produção: por IP em cadastro/login e por e-mail no login,
	// o que também freia força bruta distribuída contra uma conta só.
	registerIP, loginIP, loginEmail *httpx.Limiter
	trustProxy                      bool
}

func newAPI(st store, keys *authn.Keys, cfg Config) *api {
	dummy, _ := bcrypt.GenerateFromPassword([]byte("dummy-password"), bcrypt.DefaultCost)
	a := &api{st: st, keys: keys, secure: cfg.Secure, google: newGoogle(cfg), dummyHash: dummy, trustProxy: cfg.TrustProxy}
	if cfg.RateLimit {
		a.registerIP = httpx.NewLimiter(5, 10)
		a.loginIP = httpx.NewLimiter(10, 20)
		a.loginEmail = httpx.NewLimiter(4, 8)
	}
	return a
}

// Rotas e cookies idênticos aos dos serviços antigos, para o frontend não perceber a migração.
func (a *api) routes(mux *http.ServeMux) {
	auth := func(h http.HandlerFunc) http.Handler { return a.keys.Require(h) }

	mux.Handle("POST /user", a.limit(a.registerIP, http.HandlerFunc(a.register)))
	mux.Handle("GET /users", auth(a.listUsers))
	mux.Handle("GET /user/{id}", auth(a.getUser))
	mux.Handle("PUT /user/me", auth(a.updateMe))
	mux.Handle("DELETE /user/me", auth(a.deleteMe))

	mux.Handle("POST /auth/login", a.limit(a.loginIP, http.HandlerFunc(a.login)))
	mux.HandleFunc("POST /auth/logout", a.logout)
	mux.Handle("GET /auth/me", auth(a.me))

	mux.HandleFunc("GET /oauth/google/login", a.googleLogin)
	mux.HandleFunc("GET /oauth/google/callback", a.googleCallback)
}

// limit aplica o limitador por IP quando configurado.
func (a *api) limit(l *httpx.Limiter, h http.Handler) http.Handler {
	if l == nil {
		return h
	}
	return httpx.RateLimit(l, a.trustProxy)(h)
}

const (
	maxUsernameLen = 50
	maxEmailLen    = 254
	maxPictureLen  = 2048
)

func normEmail(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

func (a *api) register(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Picture  string `json:"picture"`
	}
	if !httpx.Decode(w, r, &in) {
		return
	}
	in.Username, in.Email = strings.TrimSpace(in.Username), normEmail(in.Email)

	switch {
	case in.Username == "" || in.Email == "" || in.Password == "":
		httpx.Error(w, http.StatusBadRequest, "username, email e password são obrigatórios")
		return
	case !strings.Contains(in.Email, "@") || len(in.Email) > maxEmailLen:
		httpx.Error(w, http.StatusBadRequest, "e-mail inválido")
		return
	case utf8.RuneCountInString(in.Username) > maxUsernameLen:
		httpx.Error(w, http.StatusBadRequest, "username deve ter no máximo 50 caracteres")
		return
	case len(in.Picture) > maxPictureLen:
		httpx.Error(w, http.StatusBadRequest, "URL da imagem muito longa")
		return
	case len(in.Password) < 8 || len(in.Password) > 72: // 72 = limite do bcrypt
		httpx.Error(w, http.StatusBadRequest, "a senha deve ter entre 8 e 72 caracteres")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		slog.ErrorContext(r.Context(), "bcrypt", "err", err)
		httpx.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}
	u, err := a.st.create(r.Context(), newUser{
		Username: in.Username, Email: in.Email, PasswordHash: string(hash), Picture: in.Picture,
	})
	switch {
	case errors.Is(err, ErrEmailTaken):
		httpx.Error(w, http.StatusConflict, "e-mail já cadastrado")
	case err != nil:
		slog.ErrorContext(r.Context(), "criar usuário", "err", err)
		httpx.Error(w, http.StatusInternalServerError, "erro ao registrar usuário")
	default:
		httpx.JSON(w, http.StatusCreated, map[string]string{"user": u.ID})
	}
}

func (a *api) login(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Type     string `json:"type"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !httpx.Decode(w, r, &in) {
		return
	}
	if in.Type != "" && in.Type != "password" {
		httpx.Error(w, http.StatusBadRequest, "método de autenticação não suportado (use o fluxo /oauth/google/login)")
		return
	}

	email := normEmail(in.Email)
	if a.loginEmail != nil && !a.loginEmail.Allow(email) {
		httpx.TooManyRequests(w)
		return
	}
	row, err := a.st.byEmail(r.Context(), email)
	if err != nil && !errors.Is(err, ErrNotFound) {
		slog.ErrorContext(r.Context(), "buscar usuário", "err", err)
		httpx.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}

	// Sempre executa um bcrypt (mesmo sem usuário) e responde igual em qualquer falha:
	// evita enumerar e-mails por mensagem ou por tempo de resposta.
	hash := a.dummyHash
	usable := err == nil && !row.IsOAuth && row.passwordHash != ""
	if usable {
		hash = []byte(row.passwordHash)
	}
	cmp := bcrypt.CompareHashAndPassword(hash, []byte(in.Password))
	if !usable || cmp != nil {
		httpx.Error(w, http.StatusUnauthorized, "credenciais inválidas")
		return
	}

	a.startSession(w, r, row.User, true)
}

// startSession emite os cookies. Com jsonBody=true responde {"user","csrf_token"} (formato antigo).
func (a *api) startSession(w http.ResponseWriter, r *http.Request, u User, jsonBody bool) (csrf string, ok bool) {
	token, err := a.keys.Sign(u.ID, u.Username)
	if err == nil {
		csrf, err = authn.NewCSRFToken()
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "emitir sessão", "err", err)
		httpx.Error(w, http.StatusInternalServerError, "erro interno")
		return "", false
	}
	authn.SetSession(w, token, csrf, a.secure)
	if jsonBody {
		httpx.JSON(w, http.StatusOK, map[string]any{"user": u, "csrf_token": csrf})
	}
	return csrf, true
}

func (a *api) logout(w http.ResponseWriter, _ *http.Request) {
	authn.ClearSession(w, a.secure)
	httpx.JSON(w, http.StatusOK, map[string]string{"message": "logout successful"})
}

func (a *api) me(w http.ResponseWriter, r *http.Request) {
	id, _ := authn.From(r.Context())
	row, err := a.st.byID(r.Context(), id.UserID)
	if errors.Is(err, ErrNotFound) {
		httpx.Error(w, http.StatusNotFound, "usuário não encontrado")
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "me", "err", err)
		httpx.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}
	httpx.JSON(w, http.StatusOK, row.User)
}

func (a *api) listUsers(w http.ResponseWriter, r *http.Request) {
	users, err := a.st.list(r.Context())
	if err != nil {
		slog.ErrorContext(r.Context(), "listar usuários", "err", err)
		httpx.Error(w, http.StatusInternalServerError, "erro ao listar usuários")
		return
	}
	httpx.JSON(w, http.StatusOK, users)
}

func (a *api) getUser(w http.ResponseWriter, r *http.Request) {
	row, err := a.st.byID(r.Context(), r.PathValue("id"))
	if errors.Is(err, ErrNotFound) {
		httpx.Error(w, http.StatusNotFound, "usuário não encontrado")
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "buscar usuário", "err", err)
		httpx.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}
	httpx.JSON(w, http.StatusOK, row.User)
}

func (a *api) updateMe(w http.ResponseWriter, r *http.Request) {
	id, _ := authn.From(r.Context())
	var in struct {
		Username string `json:"username"`
		Picture  string `json:"picture"`
	}
	if !httpx.Decode(w, r, &in) {
		return
	}
	in.Username = strings.TrimSpace(in.Username)
	if in.Username == "" {
		httpx.Error(w, http.StatusBadRequest, "username não pode ser vazio")
		return
	}
	if utf8.RuneCountInString(in.Username) > maxUsernameLen || len(in.Picture) > maxPictureLen {
		httpx.Error(w, http.StatusBadRequest, "username (máx. 50) ou URL da imagem (máx. 2048) grandes demais")
		return
	}
	u, err := a.st.update(r.Context(), id.UserID, in.Username, in.Picture)
	if errors.Is(err, ErrNotFound) {
		httpx.Error(w, http.StatusNotFound, "usuário não encontrado")
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "atualizar usuário", "err", err)
		httpx.Error(w, http.StatusInternalServerError, "erro ao atualizar usuário")
		return
	}
	httpx.JSON(w, http.StatusOK, u)
}

func (a *api) deleteMe(w http.ResponseWriter, r *http.Request) {
	id, _ := authn.From(r.Context())
	err := a.st.softDelete(r.Context(), id.UserID)
	if errors.Is(err, ErrNotFound) {
		httpx.Error(w, http.StatusNotFound, "usuário não encontrado")
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "remover usuário", "err", err)
		httpx.Error(w, http.StatusInternalServerError, "erro ao remover usuário")
		return
	}
	authn.ClearSession(w, a.secure)
	httpx.JSON(w, http.StatusOK, id.UserID)
}

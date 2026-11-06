// Package authn emite e valida a sessão (JWT RS256 em cookie HttpOnly + CSRF double-submit).
// Substitui o auth-middleware/forwardAuth: a validação agora é in-process e a chave é
// carregada uma única vez no boot (antes era relida do disco a cada requisição).
package authn

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/FelipeFelipeRenan/goverse/internal/platform/httpx"
	"github.com/golang-jwt/jwt/v5"
)

const (
	AccessCookie = "access_token"
	CSRFCookie   = "csrf_token"
	CSRFHeader   = "X-CSRF-TOKEN"
	SessionTTL   = 24 * time.Hour
)

// Claims mantém os nomes de campo do token atual (user_id, username) para não invalidar clientes.
type Claims struct {
	UserID   string `json:"user_id"`
	UserName string `json:"username"`
	jwt.RegisteredClaims
}

type Keys struct{ priv *rsa.PrivateKey }

func NewKeys(priv *rsa.PrivateKey) *Keys { return &Keys{priv: priv} }

// LoadKeys lê a chave privada (PKCS#1 ou PKCS#8). A pública é derivada dela.
func LoadKeys(path string) (*Keys, error) {
	pem, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ler chave privada: %w", err)
	}
	priv, err := jwt.ParseRSAPrivateKeyFromPEM(pem)
	if err != nil {
		return nil, fmt.Errorf("parse da chave privada: %w", err)
	}
	return NewKeys(priv), nil
}

func (k *Keys) Sign(userID, username string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		UserName: username,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(SessionTTL)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(k.priv)
}

// Verify aceita somente RS256 e exige expiração.
func (k *Keys) Verify(token string) (*Claims, error) {
	c := &Claims{}
	t, err := jwt.ParseWithClaims(token, c,
		func(*jwt.Token) (any, error) { return &k.priv.PublicKey, nil },
		jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, err
	}
	if !t.Valid || c.UserID == "" {
		return nil, errors.New("token inválido")
	}
	return c, nil
}

// Identity é o usuário autenticado da requisição.
type Identity struct {
	UserID   string
	Username string
}

type ctxKey struct{}

func WithIdentity(ctx context.Context, id Identity) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

func From(ctx context.Context) (Identity, bool) {
	id, ok := ctx.Value(ctxKey{}).(Identity)
	return id, ok
}

// Require exige sessão válida e, em métodos que alteram estado, o CSRF (cookie == header).
func (k *Keys) Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(AccessCookie)
		if err != nil {
			httpx.Error(w, http.StatusUnauthorized, "não autenticado")
			return
		}
		claims, err := k.Verify(c.Value)
		if err != nil {
			httpx.Error(w, http.StatusUnauthorized, "sessão inválida ou expirada")
			return
		}
		if unsafeMethod(r.Method) && !validCSRF(r) {
			httpx.Error(w, http.StatusForbidden, "token CSRF inválido")
			return
		}
		ctx := WithIdentity(r.Context(), Identity{UserID: claims.UserID, Username: claims.UserName})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func unsafeMethod(m string) bool {
	switch m {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	}
	return false
}

func validCSRF(r *http.Request) bool {
	c, err := r.Cookie(CSRFCookie)
	if err != nil || c.Value == "" {
		return false
	}
	h := r.Header.Get(CSRFHeader)
	return h != "" && subtle.ConstantTimeCompare([]byte(c.Value), []byte(h)) == 1
}

func NewCSRFToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// SameSite: em produção o frontend é cross-site (SameSite=None exige Secure); em dev, Lax.
func SameSite(secure bool) http.SameSite {
	if secure {
		return http.SameSiteNoneMode
	}
	return http.SameSiteLaxMode
}

func SetSession(w http.ResponseWriter, token, csrf string, secure bool) {
	maxAge := int(SessionTTL.Seconds())
	http.SetCookie(w, &http.Cookie{
		Name: AccessCookie, Value: token, MaxAge: maxAge, Path: "/",
		HttpOnly: true, Secure: secure, SameSite: SameSite(secure),
	})
	// O CSRF precisa ser legível pelo JS do frontend: não é HttpOnly.
	http.SetCookie(w, &http.Cookie{
		Name: CSRFCookie, Value: csrf, MaxAge: maxAge, Path: "/",
		Secure: secure, SameSite: SameSite(secure),
	})
}

func ClearSession(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name: AccessCookie, MaxAge: -1, Path: "/",
		HttpOnly: true, Secure: secure, SameSite: SameSite(secure),
	})
	http.SetCookie(w, &http.Cookie{
		Name: CSRFCookie, MaxAge: -1, Path: "/",
		Secure: secure, SameSite: SameSite(secure),
	})
}

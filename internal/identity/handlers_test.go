package identity

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/FelipeFelipeRenan/goverse/internal/platform/authn"
)

// fakeStore implementa store em memória: os handlers são testados sem Postgres.
type fakeStore struct {
	rows map[string]userRow
	seq  int
}

func newFakeStore() *fakeStore { return &fakeStore{rows: map[string]userRow{}} }

func (f *fakeStore) create(_ context.Context, u newUser) (User, error) {
	for _, r := range f.rows {
		if r.Email == u.Email {
			return User{}, ErrEmailTaken
		}
	}
	f.seq++
	id := fmt.Sprintf("id-%d", f.seq)
	r := userRow{
		User:         User{ID: id, Username: u.Username, Email: u.Email, Picture: u.Picture, IsOAuth: u.IsOAuth, CreatedAt: time.Now()},
		passwordHash: u.PasswordHash,
	}
	f.rows[id] = r
	return r.User, nil
}

func (f *fakeStore) byID(_ context.Context, id string) (userRow, error) {
	if r, ok := f.rows[id]; ok {
		return r, nil
	}
	return userRow{}, ErrNotFound
}

func (f *fakeStore) byEmail(_ context.Context, email string) (userRow, error) {
	for _, r := range f.rows {
		if r.Email == email {
			return r, nil
		}
	}
	return userRow{}, ErrNotFound
}

func (f *fakeStore) many(_ context.Context, ids []string) ([]User, error) {
	var out []User
	for _, id := range ids {
		if r, ok := f.rows[id]; ok {
			out = append(out, r.User)
		}
	}
	return out, nil
}

func (f *fakeStore) list(_ context.Context) ([]User, error) {
	out := []User{}
	for _, r := range f.rows {
		out = append(out, r.User)
	}
	return out, nil
}

func (f *fakeStore) update(_ context.Context, id, username, picture string) (User, error) {
	r, ok := f.rows[id]
	if !ok {
		return User{}, ErrNotFound
	}
	r.Username = username
	if picture != "" {
		r.Picture = picture
	}
	f.rows[id] = r
	return r.User, nil
}

func (f *fakeStore) softDelete(_ context.Context, id string) error {
	if _, ok := f.rows[id]; !ok {
		return ErrNotFound
	}
	delete(f.rows, id)
	return nil
}

func newTestServer(t *testing.T) (http.Handler, *fakeStore) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	st := newFakeStore()
	mux := http.NewServeMux()
	newAPI(st, authn.NewKeys(priv), Config{}).routes(mux)
	return mux, st
}

func do(h http.Handler, method, path, body string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func register(t *testing.T, h http.Handler, email, pass string) {
	t.Helper()
	rec := do(h, "POST", "/user", fmt.Sprintf(`{"username":"Ana","email":%q,"password":%q}`, email, pass))
	if rec.Code != http.StatusCreated {
		t.Fatalf("register: %d %s", rec.Code, rec.Body)
	}
}

func login(t *testing.T, h http.Handler, email, pass string) []*http.Cookie {
	t.Helper()
	rec := do(h, "POST", "/auth/login", fmt.Sprintf(`{"type":"password","email":%q,"password":%q}`, email, pass))
	if rec.Code != http.StatusOK {
		t.Fatalf("login: %d %s", rec.Code, rec.Body)
	}
	return rec.Result().Cookies()
}

func cookie(cs []*http.Cookie, name string) *http.Cookie {
	for _, c := range cs {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func TestRegisterAndLoginSetCookies(t *testing.T) {
	h, _ := newTestServer(t)
	register(t, h, "Ana@Example.com", "senha1234") // e-mail é normalizado para minúsculas

	rec := do(h, "POST", "/auth/login", `{"type":"password","email":"ana@example.com","password":"senha1234"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("login: %d %s", rec.Code, rec.Body)
	}
	access, csrf := cookie(rec.Result().Cookies(), authn.AccessCookie), cookie(rec.Result().Cookies(), authn.CSRFCookie)
	if access == nil || !access.HttpOnly {
		t.Fatalf("access_token ausente ou sem HttpOnly: %+v", access)
	}
	if csrf == nil || csrf.HttpOnly {
		t.Fatalf("csrf_token ausente ou HttpOnly (o JS precisa ler): %+v", csrf)
	}
	var body struct {
		CSRF string `json:"csrf_token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body.CSRF != csrf.Value {
		t.Fatalf("csrf_token do corpo deve igualar o cookie: %v %q", err, body.CSRF)
	}
}

func TestRegisterValidationAndDuplicate(t *testing.T) {
	h, _ := newTestServer(t)
	if rec := do(h, "POST", "/user", `{"username":"a","email":"a@b.c","password":"curta"}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("senha curta: %d", rec.Code)
	}
	register(t, h, "a@b.c", "senha1234")
	if rec := do(h, "POST", "/user", `{"username":"b","email":"A@B.C","password":"senha1234"}`); rec.Code != http.StatusConflict {
		t.Fatalf("duplicado deveria ser 409, veio %d", rec.Code)
	}
}

func TestLoginFailuresAreIndistinguishable(t *testing.T) {
	h, _ := newTestServer(t)
	register(t, h, "a@b.c", "senha1234")

	wrongPass := do(h, "POST", "/auth/login", `{"email":"a@b.c","password":"errada123"}`)
	noUser := do(h, "POST", "/auth/login", `{"email":"x@y.z","password":"senha1234"}`)
	if wrongPass.Code != http.StatusUnauthorized || noUser.Code != http.StatusUnauthorized {
		t.Fatalf("esperava 401/401, veio %d/%d", wrongPass.Code, noUser.Code)
	}
	if wrongPass.Body.String() != noUser.Body.String() {
		t.Fatalf("mensagens diferentes permitem enumerar e-mails:\n%s\n%s", wrongPass.Body, noUser.Body)
	}
}

func TestOAuthOnlyAccountCannotUsePasswordLogin(t *testing.T) {
	h, st := newTestServer(t)
	if _, err := st.create(context.Background(), newUser{Username: "G", Email: "g@x.com", IsOAuth: true}); err != nil {
		t.Fatal(err)
	}
	if rec := do(h, "POST", "/auth/login", `{"email":"g@x.com","password":""}`); rec.Code != http.StatusUnauthorized {
		t.Fatalf("conta OAuth sem senha deveria dar 401, veio %d", rec.Code)
	}
}

func TestUsersRequireAuthAndNeverLeakPassword(t *testing.T) {
	h, _ := newTestServer(t)
	register(t, h, "a@b.c", "senha1234")

	if rec := do(h, "GET", "/users", ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("/users sem sessão deveria dar 401, veio %d", rec.Code)
	}
	cs := login(t, h, "a@b.c", "senha1234")
	rec := do(h, "GET", "/users", "", cs...)
	if rec.Code != http.StatusOK {
		t.Fatalf("/users: %d %s", rec.Code, rec.Body)
	}
	if body := strings.ToLower(rec.Body.String()); strings.Contains(body, "password") || strings.Contains(body, "$2a$") {
		t.Fatalf("resposta vaza senha/hash: %s", rec.Body)
	}
}

func TestCSRFEnforcedOnStateChangingRoutes(t *testing.T) {
	h, _ := newTestServer(t)
	register(t, h, "a@b.c", "senha1234")
	cs := login(t, h, "a@b.c", "senha1234")
	csrf := cookie(cs, authn.CSRFCookie).Value

	if rec := do(h, "PUT", "/user/me", `{"username":"Novo"}`, cs...); rec.Code != http.StatusForbidden {
		t.Fatalf("sem header CSRF deveria dar 403, veio %d", rec.Code)
	}

	req := httptest.NewRequest("PUT", "/user/me", strings.NewReader(`{"username":"Novo"}`))
	for _, c := range cs {
		req.AddCookie(c)
	}
	req.Header.Set(authn.CSRFHeader, csrf)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Novo") {
		t.Fatalf("update com CSRF: %d %s", rec.Code, rec.Body)
	}
}

func TestMeAndLogout(t *testing.T) {
	h, _ := newTestServer(t)
	register(t, h, "a@b.c", "senha1234")
	cs := login(t, h, "a@b.c", "senha1234")

	if rec := do(h, "GET", "/auth/me", "", cs...); rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "a@b.c") {
		t.Fatalf("/auth/me: %d %s", rec.Code, rec.Body)
	}
	rec := do(h, "POST", "/auth/logout", "")
	if c := cookie(rec.Result().Cookies(), authn.AccessCookie); c == nil || c.MaxAge >= 0 {
		t.Fatalf("logout deve expirar o access_token: %+v", c)
	}
}

func TestDirectory(t *testing.T) {
	_, st := newTestServer(t)
	u, _ := st.create(context.Background(), newUser{Username: "A", Email: "a@b.c"})
	d := directory{st}

	if ok, err := d.Exists(context.Background(), u.ID); err != nil || !ok {
		t.Fatalf("Exists: %v %v", ok, err)
	}
	if ok, err := d.Exists(context.Background(), "nao-existe"); err != nil || ok {
		t.Fatalf("Exists(inexistente): %v %v", ok, err)
	}
	m, err := d.GetMany(context.Background(), []string{u.ID, "nao-existe"})
	if err != nil || len(m) != 1 || m[u.ID].Email != "a@b.c" {
		t.Fatalf("GetMany: %v %v", m, err)
	}
}

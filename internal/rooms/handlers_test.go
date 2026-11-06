package rooms

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/FelipeFelipeRenan/goverse/internal/identity"
	"github.com/FelipeFelipeRenan/goverse/internal/platform/authn"
)

// ---- fakes em memória: os handlers e o serviço são testados sem Postgres ----

type fakeStore struct {
	rooms map[string]*Room
	mem   map[string]map[string]Member
	seq   int
}

func newFakeStore() *fakeStore {
	return &fakeStore{rooms: map[string]*Room{}, mem: map[string]map[string]Member{}}
}

func (f *fakeStore) createRoom(_ context.Context, n newRoom) (Room, error) {
	f.seq++
	id := fmt.Sprintf("room-%d", f.seq)
	now := time.Now()
	r := &Room{ID: id, Name: n.Name, Description: n.Description, IsPublic: n.IsPublic, OwnerID: n.OwnerID,
		MemberCount: 1, MaxMembers: n.MaxMembers, CreatedAt: now.Add(time.Duration(f.seq) * time.Second), UpdatedAt: now}
	f.rooms[id] = r
	f.mem[id] = map[string]Member{n.OwnerID: {RoomID: id, UserID: n.OwnerID, Role: RoleOwner, JoinedAt: now}}
	return *r, nil
}

func (f *fakeStore) room(_ context.Context, id string) (Room, error) {
	if r, ok := f.rooms[id]; ok {
		return *r, nil
	}
	return Room{}, ErrRoomNotFound
}

func (f *fakeStore) sorted(keep func(*Room) bool) []Room {
	out := []Room{}
	for _, r := range f.rooms {
		if keep(r) {
			out = append(out, *r)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

func (f *fakeStore) listPublic(_ context.Context, limit, offset int, keyword string) ([]Room, error) {
	all := f.sorted(func(r *Room) bool {
		return r.IsPublic && strings.Contains(strings.ToLower(r.Name), strings.ToLower(keyword))
	})
	if offset >= len(all) {
		return []Room{}, nil
	}
	all = all[offset:]
	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

func (f *fakeStore) roomsOf(_ context.Context, userID string) ([]Room, error) {
	return f.sorted(func(r *Room) bool { _, ok := f.mem[r.ID][userID]; return ok }), nil
}

func (f *fakeStore) ownedBy(_ context.Context, userID string) ([]Room, error) {
	return f.sorted(func(r *Room) bool { return r.OwnerID == userID }), nil
}

func (f *fakeStore) update(_ context.Context, id string, p roomPatch) (Room, error) {
	r, ok := f.rooms[id]
	if !ok {
		return Room{}, ErrRoomNotFound
	}
	if p.MaxMembers != nil && *p.MaxMembers < r.MemberCount {
		return Room{}, invalid("valor inválido: max_members não pode ser menor que o número de membros")
	}
	if p.Name != nil {
		r.Name = *p.Name
	}
	if p.Description != nil {
		r.Description = *p.Description
	}
	if p.IsPublic != nil {
		r.IsPublic = *p.IsPublic
	}
	if p.MaxMembers != nil {
		r.MaxMembers = *p.MaxMembers
	}
	return *r, nil
}

func (f *fakeStore) softDelete(_ context.Context, id string) error {
	if _, ok := f.rooms[id]; !ok {
		return ErrRoomNotFound
	}
	delete(f.rooms, id)
	return nil
}

func (f *fakeStore) member(_ context.Context, roomID, userID string) (Member, error) {
	if _, ok := f.rooms[roomID]; !ok {
		return Member{}, ErrMemberNotFound
	}
	if m, ok := f.mem[roomID][userID]; ok {
		return m, nil
	}
	return Member{}, ErrMemberNotFound
}

func (f *fakeStore) members_(roomID string) []Member {
	out := []Member{}
	for _, m := range f.mem[roomID] {
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UserID < out[j].UserID })
	return out
}

func (f *fakeStore) members(_ context.Context, roomID string) ([]Member, error) {
	return f.members_(roomID), nil
}

func (f *fakeStore) addMember(_ context.Context, roomID, userID string, role Role) error {
	r, ok := f.rooms[roomID]
	if !ok {
		return ErrRoomNotFound
	}
	if r.MemberCount >= r.MaxMembers {
		return ErrRoomFull
	}
	if _, dup := f.mem[roomID][userID]; dup {
		return ErrMemberAlreadyExists
	}
	f.mem[roomID][userID] = Member{RoomID: roomID, UserID: userID, Role: role, JoinedAt: time.Now()}
	r.MemberCount++
	return nil
}

func (f *fakeStore) removeMember(_ context.Context, roomID, userID string) error {
	if _, ok := f.mem[roomID][userID]; !ok {
		return ErrMemberNotFound
	}
	delete(f.mem[roomID], userID)
	f.rooms[roomID].MemberCount--
	return nil
}

func (f *fakeStore) setRole(_ context.Context, roomID, userID string, role Role) error {
	m, ok := f.mem[roomID][userID]
	if !ok {
		return ErrMemberNotFound
	}
	m.Role = role
	f.mem[roomID][userID] = m
	return nil
}

type fakeDir map[string]identity.User

func (d fakeDir) Exists(_ context.Context, id string) (bool, error) { _, ok := d[id]; return ok, nil }

func (d fakeDir) Get(_ context.Context, id string) (identity.User, error) {
	if u, ok := d[id]; ok {
		return u, nil
	}
	return identity.User{}, identity.ErrNotFound
}

func (d fakeDir) GetMany(_ context.Context, ids []string) (map[string]identity.User, error) {
	out := map[string]identity.User{}
	for _, id := range ids {
		if u, ok := d[id]; ok {
			out[id] = u
		}
	}
	return out, nil
}

// ---- harness ----

type env struct {
	h    http.Handler
	keys *authn.Keys
	st   *fakeStore
	svc  *service
}

func newEnv(t *testing.T, userIDs ...string) *env {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	dir := fakeDir{}
	for _, id := range userIDs {
		dir[id] = identity.User{ID: id, Username: "user-" + id, Email: id + "@x.com", CreatedAt: time.Now()}
	}
	e := &env{keys: authn.NewKeys(priv), st: newFakeStore()}
	e.svc = &service{st: e.st, users: dir}
	mux := http.NewServeMux()
	(&api{svc: e.svc, keys: e.keys}).routes(mux)
	e.h = mux
	return e
}

// call faz a requisição como userID (sessão + CSRF válidos); userID "" = anônimo.
func (e *env) call(userID, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if userID != "" {
		tok, _ := e.keys.Sign(userID, "user-"+userID)
		req.AddCookie(&http.Cookie{Name: authn.AccessCookie, Value: tok})
		req.AddCookie(&http.Cookie{Name: authn.CSRFCookie, Value: "csrf"})
		req.Header.Set(authn.CSRFHeader, "csrf")
	}
	rec := httptest.NewRecorder()
	e.h.ServeHTTP(rec, req)
	return rec
}

func (e *env) mustCreate(t *testing.T, owner, body string) string {
	t.Helper()
	rec := e.call(owner, "POST", "/rooms", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("criar sala: %d %s", rec.Code, rec.Body)
	}
	var r struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &r); err != nil || r.ID == "" {
		t.Fatalf("resposta sem id: %v %s", err, rec.Body)
	}
	return r.ID
}

func want(t *testing.T, rec *httptest.ResponseRecorder, code int) {
	t.Helper()
	if rec.Code != code {
		t.Fatalf("esperava %d, veio %d: %s", code, rec.Code, rec.Body)
	}
}

// ---- testes ----

func TestRequiresAuth(t *testing.T) {
	e := newEnv(t, "a")
	want(t, e.call("", "POST", "/rooms", `{"name":"x"}`), http.StatusUnauthorized)
	want(t, e.call("", "GET", "/rooms/mine", ""), http.StatusUnauthorized)
	want(t, e.call("", "GET", "/rooms", ""), http.StatusOK) // única rota aberta
}

func TestCreateDefaultsAndResponseShapes(t *testing.T) {
	e := newEnv(t, "a")
	id := e.mustCreate(t, "a", `{"name":"  Sala A  ","is_public":true}`)

	rec := e.call("a", "GET", "/rooms/"+id, "")
	want(t, rec, http.StatusOK)
	var one map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &one)
	if one["id"] != id || one["name"] != "Sala A" || one["max_members"] != float64(10) || one["member_count"] != float64(1) {
		t.Fatalf("GET /rooms/{id} (formato id/name): %v", one)
	}

	rec = e.call("a", "GET", "/rooms/mine", "")
	want(t, rec, http.StatusOK)
	var list []map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &list)
	if len(list) != 1 || list[0]["room_id"] != id || list[0]["room_name"] != "Sala A" {
		t.Fatalf("listagem deve manter o formato legado room_id/room_name: %v", list)
	}

	want(t, e.call("a", "POST", "/rooms", `{"name":"   "}`), http.StatusBadRequest)
	want(t, e.call("a", "POST", "/rooms", `{"name":"x","max_members":5000}`), http.StatusBadRequest)
}

func TestSessionOfDeletedUserCannotCreateRoom(t *testing.T) {
	e := newEnv(t, "a")
	want(t, e.call("fantasma", "POST", "/rooms", `{"name":"x"}`), http.StatusUnauthorized)
}

func TestPublicListNeverShowsPrivateRooms(t *testing.T) {
	e := newEnv(t, "a", "b")
	pub := e.mustCreate(t, "a", `{"name":"publica","is_public":true}`)
	e.mustCreate(t, "b", `{"name":"secreta","is_public":false}`)

	for _, q := range []string{"/rooms", "/rooms?public_only=false", "/rooms?keyword=secreta", "/rooms?keyword=sec"} {
		rec := e.call("", "GET", q, "")
		want(t, rec, http.StatusOK)
		if strings.Contains(rec.Body.String(), "secreta") {
			t.Fatalf("%s vazou sala privada: %s", q, rec.Body)
		}
	}
	if rec := e.call("", "GET", "/rooms", ""); !strings.Contains(rec.Body.String(), pub) {
		t.Fatalf("sala pública deveria aparecer: %s", rec.Body)
	}
	// vazio => [] (nunca null)
	if rec := e.call("", "GET", "/rooms?keyword=nada-assim", ""); strings.TrimSpace(rec.Body.String()) != "[]" {
		t.Fatalf("lista vazia deveria ser []: %q", rec.Body)
	}
}

func TestPrivateRoomInvisibleToNonMembers(t *testing.T) {
	e := newEnv(t, "a", "b")
	id := e.mustCreate(t, "a", `{"name":"secreta","is_public":false}`)

	want(t, e.call("b", "GET", "/rooms/"+id, ""), http.StatusNotFound)
	want(t, e.call("b", "GET", "/rooms/"+id+"/members", ""), http.StatusNotFound)
	want(t, e.call("b", "POST", "/rooms/"+id+"/join", ""), http.StatusForbidden)
	want(t, e.call("a", "GET", "/rooms/"+id, ""), http.StatusOK)
}

func TestJoinIsIdempotentAndRoomFullIs409(t *testing.T) {
	e := newEnv(t, "a", "b", "c")
	id := e.mustCreate(t, "a", `{"name":"dupla","is_public":true,"max_members":2}`)

	want(t, e.call("b", "POST", "/rooms/"+id+"/join", ""), http.StatusOK)
	want(t, e.call("b", "POST", "/rooms/"+id+"/join", ""), http.StatusOK) // já é membro
	if r, _ := e.st.room(context.Background(), id); r.MemberCount != 2 {
		t.Fatalf("member_count = %d, esperava 2", r.MemberCount)
	}
	want(t, e.call("c", "POST", "/rooms/"+id+"/join", ""), http.StatusConflict) // lotada
}

func TestOnlyOwnerOrAdminCanEditAndOnlyOwnerCanDelete(t *testing.T) {
	e := newEnv(t, "a", "b", "c")
	id := e.mustCreate(t, "a", `{"name":"x","is_public":true,"max_members":5}`)
	want(t, e.call("b", "POST", "/rooms/"+id+"/join", ""), http.StatusOK)

	want(t, e.call("b", "PATCH", "/rooms/"+id, `{"name":"hack"}`), http.StatusForbidden) // membro comum
	want(t, e.call("c", "PATCH", "/rooms/"+id, `{"name":"hack"}`), http.StatusForbidden) // nem é membro
	want(t, e.call("b", "DELETE", "/rooms/"+id, ""), http.StatusForbidden)

	want(t, e.call("a", "PUT", "/rooms/"+id+"/members/b/role", `{"new_role":"admin"}`), http.StatusOK)
	rec := e.call("b", "PATCH", "/rooms/"+id, `{"name":"Novo","is_public":false,"max_members":3}`)
	want(t, rec, http.StatusOK)
	var got map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &got)
	if got["room_name"] != "Novo" || got["is_public"] != false || got["max_members"] != float64(3) {
		t.Fatalf("PATCH deve aplicar todos os campos (antes ignorava is_public/max_members): %v", got)
	}
	want(t, e.call("a", "PATCH", "/rooms/"+id, `{"max_members":1}`), http.StatusBadRequest) // menor que membros atuais

	want(t, e.call("b", "DELETE", "/rooms/"+id, ""), http.StatusForbidden) // admin não apaga
	want(t, e.call("a", "DELETE", "/rooms/"+id, ""), http.StatusOK)
	want(t, e.call("a", "GET", "/rooms/"+id, ""), http.StatusNotFound)
}

func TestAddMemberRules(t *testing.T) {
	e := newEnv(t, "a", "b", "c")
	id := e.mustCreate(t, "a", `{"name":"x","is_public":true,"max_members":5}`)
	path := "/rooms/" + id + "/members"

	want(t, e.call("b", "POST", path, `{"user_id":"c"}`), http.StatusForbidden)                 // só o dono adiciona
	want(t, e.call("a", "POST", path, `{"user_id":"c","role":"owner"}`), http.StatusBadRequest) // não vira dono
	want(t, e.call("a", "POST", path, `{"user_id":"c","role":"deus"}`), http.StatusBadRequest)
	want(t, e.call("a", "POST", path, `{"user_id":"fantasma"}`), http.StatusNotFound)
	want(t, e.call("a", "POST", path, `{"user_id":"c","role":"admin"}`), http.StatusCreated)
	want(t, e.call("a", "POST", path, `{"user_id":"c"}`), http.StatusConflict) // já é membro
	if m, _ := e.st.member(context.Background(), id, "c"); m.Role != RoleAdmin {
		t.Fatalf("role = %q", m.Role)
	}
}

func TestRemoveMemberRules(t *testing.T) {
	e := newEnv(t, "a", "b", "c")
	id := e.mustCreate(t, "a", `{"name":"x","is_public":true,"max_members":5}`)
	want(t, e.call("b", "POST", "/rooms/"+id+"/join", ""), http.StatusOK)
	want(t, e.call("c", "POST", "/rooms/"+id+"/join", ""), http.StatusOK)

	want(t, e.call("a", "DELETE", "/rooms/"+id+"/members/a", ""), http.StatusConflict) // dono não sai
	want(t, e.call("b", "DELETE", "/rooms/"+id+"/members/c", ""), http.StatusForbidden)
	want(t, e.call("b", "DELETE", "/rooms/"+id+"/members/b", ""), http.StatusOK) // sai sozinho
	want(t, e.call("a", "DELETE", "/rooms/"+id+"/members/c", ""), http.StatusOK) // dono expulsa
	if r, _ := e.st.room(context.Background(), id); r.MemberCount != 1 {
		t.Fatalf("member_count = %d, esperava 1", r.MemberCount)
	}
}

func TestSetRoleRules(t *testing.T) {
	e := newEnv(t, "a", "b")
	id := e.mustCreate(t, "a", `{"name":"x","is_public":true}`)
	want(t, e.call("b", "POST", "/rooms/"+id+"/join", ""), http.StatusOK)

	p := func(u string) string { return "/rooms/" + id + "/members/" + u + "/role" }
	want(t, e.call("b", "PUT", p("b"), `{"new_role":"admin"}`), http.StatusForbidden) // não se promove
	want(t, e.call("a", "PUT", p("a"), `{"new_role":"member"}`), http.StatusConflict) // dono intocável
	want(t, e.call("a", "PUT", p("b"), `{"new_role":"owner"}`), http.StatusBadRequest)
	want(t, e.call("a", "PUT", p("b"), `{"new_role":"deus"}`), http.StatusBadRequest)
	want(t, e.call("a", "PUT", p("fantasma"), `{"new_role":"admin"}`), http.StatusNotFound)
	want(t, e.call("a", "PUT", p("b"), `{"new_role":"moderator"}`), http.StatusOK)
}

func TestMembersAreEnrichedWithUsersAndSkipMissing(t *testing.T) {
	e := newEnv(t, "a", "b")
	id := e.mustCreate(t, "a", `{"name":"x","is_public":true}`)
	want(t, e.call("b", "POST", "/rooms/"+id+"/join", ""), http.StatusOK)

	rec := e.call("a", "GET", "/rooms/"+id+"/members", "")
	want(t, rec, http.StatusOK)
	var ms []struct {
		Role string `json:"role"`
		User struct {
			ID    string `json:"user_id"`
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"user"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &ms); err != nil || len(ms) != 2 {
		t.Fatalf("members: %v %s", err, rec.Body)
	}
	if ms[0].User.Email != "a@x.com" || ms[0].User.Name != "user-a" || ms[1].User.ID != "b" {
		t.Fatalf("usuários não foram resolvidos via Directory: %+v", ms)
	}

	// usuário removido da identity some da lista (comportamento do serviço antigo)
	delete(e.svc.users.(fakeDir), "b")
	rec = e.call("a", "GET", "/rooms/"+id+"/members", "")
	_ = json.Unmarshal(rec.Body.Bytes(), &ms)
	if len(ms) != 1 || ms[0].User.ID != "a" {
		t.Fatalf("membro sem usuário deveria sumir: %s", rec.Body)
	}
}

func TestMembershipContract(t *testing.T) {
	e := newEnv(t, "a", "b")
	id := e.mustCreate(t, "a", `{"name":"x","is_public":true}`)
	var m Membership = e.svc

	if ok, err := m.IsMember(context.Background(), id, "a"); err != nil || !ok {
		t.Fatalf("dono deveria ser membro: %v %v", ok, err)
	}
	if ok, err := m.IsMember(context.Background(), id, "b"); err != nil || ok {
		t.Fatalf("b não é membro: %v %v", ok, err)
	}
	if ok, err := m.IsMember(context.Background(), "sala-inexistente", "a"); err != nil || ok {
		t.Fatalf("sala inexistente: %v %v", ok, err)
	}
}

func TestMemberEmailsOnlyVisibleToOwnerAndAdmins(t *testing.T) {
	e := newEnv(t, "a", "b", "c")
	id := e.mustCreate(t, "a", `{"name":"x","is_public":true,"max_members":5}`)
	want(t, e.call("b", "POST", "/rooms/"+id+"/join", ""), http.StatusOK)
	want(t, e.call("c", "POST", "/rooms/"+id+"/join", ""), http.StatusOK)

	emails := func(as string) map[string]string {
		rec := e.call(as, "GET", "/rooms/"+id+"/members", "")
		want(t, rec, http.StatusOK)
		var ms []struct {
			User struct {
				ID    string `json:"user_id"`
				Email string `json:"email"`
			} `json:"user"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &ms); err != nil {
			t.Fatal(err, rec.Body)
		}
		out := map[string]string{}
		for _, m := range ms {
			out[m.User.ID] = m.User.Email
		}
		return out
	}

	if got := emails("a"); got["a"] == "" || got["b"] == "" || got["c"] == "" {
		t.Fatalf("dono deve ver todos os e-mails: %v", got)
	}
	got := emails("b") // membro comum
	if got["b"] != "b@x.com" || got["a"] != "" || got["c"] != "" {
		t.Fatalf("membro vê só o próprio e-mail: %v", got)
	}
	want(t, e.call("a", "PUT", "/rooms/"+id+"/members/b/role", `{"new_role":"admin"}`), http.StatusOK)
	if got := emails("b"); got["a"] == "" || got["c"] == "" {
		t.Fatalf("admin deve ver os e-mails: %v", got)
	}
}

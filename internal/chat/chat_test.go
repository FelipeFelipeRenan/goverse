package chat

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/FelipeFelipeRenan/goverse/internal/platform/authn"
	"github.com/FelipeFelipeRenan/goverse/internal/platform/httpx"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type fakeStore struct {
	mu   sync.Mutex
	msgs []Message
	fail bool
}

func (f *fakeStore) save(_ context.Context, m *Message) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.fail {
		return errors.New("db fora")
	}
	id, _ := uuid.NewV7()
	m.ID, m.CreatedAt = id.String(), time.Now().UTC()
	f.msgs = append(f.msgs, *m)
	return nil
}

func (f *fakeStore) history(_ context.Context, roomID string, limit, offset int) ([]Message, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := []Message{}
	for _, m := range f.msgs {
		if m.RoomID == roomID {
			out = append(out, m)
		}
	}
	return out, nil
}

func (f *fakeStore) count() int { f.mu.Lock(); defer f.mu.Unlock(); return len(f.msgs) }

type fakeMembers map[string]bool // "room|user"

func (f fakeMembers) IsMember(_ context.Context, room, user string) (bool, error) {
	return f[room+"|"+user], nil
}

type env struct {
	srv   *httptest.Server
	keys  *authn.Keys
	st    *fakeStore
	room  string
	alice string
	bob   string
	eve   string
}

func newEnv(t *testing.T) *env {
	t.Helper()
	pk, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	keys := authn.NewKeys(pk)
	e := &env{keys: keys, st: &fakeStore{}, room: uuid.NewString(), alice: uuid.NewString(), bob: uuid.NewString(), eve: uuid.NewString()}
	members := fakeMembers{e.room + "|" + e.alice: true, e.room + "|" + e.bob: true}
	m := newModule(nil, e.st, keys, members, Config{AllowedOrigins: []string{"http://localhost:4200"}})
	mux := http.NewServeMux()
	m.Register(mux)
	// Wrap completo: garante que o AccessLog não quebra o upgrade (Hijack).
	e.srv = httptest.NewServer(httpx.Wrap(mux))
	t.Cleanup(e.srv.Close)
	return e
}

func (e *env) header(t *testing.T, userID, name, origin string) http.Header {
	t.Helper()
	tok, err := e.keys.Sign(userID, name)
	if err != nil {
		t.Fatal(err)
	}
	h := http.Header{}
	h.Set("Cookie", authn.AccessCookie+"="+tok)
	if origin != "" {
		h.Set("Origin", origin)
	}
	return h
}

func (e *env) dial(t *testing.T, userID, name string) *websocket.Conn {
	t.Helper()
	url := "ws" + strings.TrimPrefix(e.srv.URL, "http") + "/ws?roomId=" + e.room
	c, resp, err := websocket.DefaultDialer.Dial(url, e.header(t, userID, name, "http://localhost:4200"))
	if err != nil {
		t.Fatalf("dial: %v (resp=%v)", err, resp)
	}
	t.Cleanup(func() { c.Close() })
	return c
}

func read(t *testing.T, c *websocket.Conn) Message {
	t.Helper()
	_ = c.SetReadDeadline(time.Now().Add(3 * time.Second))
	var m Message
	if err := c.ReadJSON(&m); err != nil {
		t.Fatalf("read: %v", err)
	}
	return m
}

// readType pula frames até achar o tipo (presença de outros pode intercalar).
func readType(t *testing.T, c *websocket.Conn, typ string) Message {
	t.Helper()
	for i := 0; i < 10; i++ {
		if m := read(t, c); m.Type == typ {
			return m
		}
	}
	t.Fatalf("não recebi %s", typ)
	return Message{}
}

func TestChatPersistsThenBroadcasts(t *testing.T) {
	e := newEnv(t)
	a, b := e.dial(t, e.alice, "alice"), e.dial(t, e.bob, "bob")
	readType(t, a, TypePresence) // própria entrada
	readType(t, b, TypePresence)

	// O cliente tenta forjar identidade e ID: o servidor ignora.
	if err := a.WriteJSON(map[string]any{"type": "CHAT", "content": "oi", "user_id": e.bob, "username": "bob", "id": "forjado", "room_id": "outra"}); err != nil {
		t.Fatal(err)
	}
	for _, c := range []*websocket.Conn{a, b} {
		m := readType(t, c, TypeChat)
		if m.Content != "oi" || m.UserID != e.alice || m.Username != "alice" || m.RoomID != e.room {
			t.Fatalf("mensagem inesperada: %+v", m)
		}
		if _, err := uuid.Parse(m.ID); err != nil || m.ID == "forjado" {
			t.Fatalf("id deve ser gerado pelo servidor: %q", m.ID)
		}
	}
	if e.st.count() != 1 {
		t.Fatalf("esperava 1 mensagem persistida, tem %d", e.st.count())
	}
}

func TestSaveFailureDoesNotBroadcast(t *testing.T) {
	e := newEnv(t)
	e.st.fail = true
	a, b := e.dial(t, e.alice, "alice"), e.dial(t, e.bob, "bob")
	readType(t, a, TypePresence)
	readType(t, b, TypePresence)
	_ = a.WriteJSON(map[string]any{"type": "CHAT", "content": "x"})
	if m := readType(t, a, "ERROR"); m.Content == "" {
		t.Fatal("remetente deve receber erro")
	}
	_ = b.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
	var m Message
	if err := b.ReadJSON(&m); err == nil {
		t.Fatalf("bob não deveria receber nada, recebeu %+v", m)
	}
}

func TestRejectsNonMemberAndBadOrigin(t *testing.T) {
	e := newEnv(t)
	base := "ws" + strings.TrimPrefix(e.srv.URL, "http") + "/ws?roomId=" + e.room

	_, resp, err := websocket.DefaultDialer.Dial(base, e.header(t, e.eve, "eve", "http://localhost:4200"))
	if err == nil || resp == nil || resp.StatusCode != http.StatusForbidden {
		t.Fatalf("não-membro deveria levar 403, resp=%v err=%v", resp, err)
	}
	_, resp, err = websocket.DefaultDialer.Dial(base, e.header(t, e.alice, "alice", "http://evil.example"))
	if err == nil || resp == nil || resp.StatusCode != http.StatusForbidden {
		t.Fatalf("origem fora da lista deveria levar 403, resp=%v err=%v", resp, err)
	}
	_, resp, err = websocket.DefaultDialer.Dial(base, http.Header{"Origin": {"http://localhost:4200"}})
	if err == nil || resp == nil || resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("sem sessão deveria levar 401, resp=%v err=%v", resp, err)
	}
	_, resp, _ = websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(e.srv.URL, "http")+"/ws?roomId=nao-uuid", e.header(t, e.alice, "alice", ""))
	if resp == nil || resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("roomId inválido deveria levar 400, resp=%v", resp)
	}
}

func TestPresenceOnlyOnFirstAndLastTab(t *testing.T) {
	e := newEnv(t)
	b := e.dial(t, e.bob, "bob")
	readType(t, b, TypePresence)

	a1 := e.dial(t, e.alice, "alice")
	if m := readType(t, b, TypePresence); m.Content != "entrou na sala" || m.UserID != e.alice {
		t.Fatalf("esperava entrada da alice, veio %+v", m)
	}
	a2 := e.dial(t, e.alice, "alice") // 2ª aba: sem nova presença
	a1.Close()                        // ainda resta uma aba: sem "saiu"
	_ = a2.WriteJSON(map[string]any{"type": "CHAT", "content": "sigo aqui"})
	if m := read(t, b); m.Type != TypeChat {
		t.Fatalf("esperava CHAT (sem presença extra), veio %+v", m)
	}
	a2.Close()
	if m := readType(t, b, TypePresence); m.Content != "saiu da sala" {
		t.Fatalf("esperava saída, veio %+v", m)
	}
}

func TestTypingSkipsSenderAndDirectIsTargeted(t *testing.T) {
	e := newEnv(t)
	a, b := e.dial(t, e.alice, "alice"), e.dial(t, e.bob, "bob")
	readType(t, a, TypePresence)
	readType(t, b, TypePresence)

	_ = a.WriteJSON(map[string]any{"type": TypeTypingStart})
	if m := readType(t, b, TypeTypingStart); m.UserID != e.alice {
		t.Fatalf("typing deveria vir da alice: %+v", m)
	}
	_ = a.WriteJSON(map[string]any{"type": TypeSignal, "target_user_id": e.bob, "content": "sdp"})
	if m := readType(t, b, TypeSignal); m.Content != "sdp" {
		t.Fatalf("bob deveria receber o sinal: %+v", m)
	}
	if e.st.count() != 0 {
		t.Fatal("typing/sinal não devem ser persistidos")
	}
}

func TestSlowConsumerIsKickedWithoutBlockingRoom(t *testing.T) {
	st := &fakeStore{}
	h := newHub(st)
	room := uuid.NewString()
	kicked := make(chan struct{})
	slow := newConn(room, "slow", "slow", 1, func() { close(kicked) })
	fast := newConn(room, "fast", "fast", 100, func() {})
	h.join(slow)
	h.join(fast)
	for i := 0; i < 5; i++ {
		h.handle(fast, []byte(fmt.Sprintf(`{"type":"CHAT","content":"m%d"}`, i)))
	}
	select {
	case <-kicked:
	case <-time.After(time.Second):
		t.Fatal("consumidor lento deveria ser derrubado")
	}
	if st.count() != 5 {
		t.Fatalf("a sala não pode travar: %d persistidas", st.count())
	}
}

func TestHistoryRequiresMembership(t *testing.T) {
	e := newEnv(t)
	_ = e.st.save(context.Background(), &Message{RoomID: e.room, UserID: e.alice, Username: "alice", Content: "antiga", Type: TypeChat})

	get := func(userID string) (*http.Response, []Message) {
		req, _ := http.NewRequest("GET", e.srv.URL+"/rooms/"+e.room+"/messages?limit=10", nil)
		req.Header = e.header(t, userID, "u", "")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		var msgs []Message
		_ = json.NewDecoder(resp.Body).Decode(&msgs)
		return resp, msgs
	}
	if resp, _ := get(e.eve); resp.StatusCode != http.StatusForbidden {
		t.Fatalf("não-membro: %d", resp.StatusCode)
	}
	resp, msgs := get(e.alice)
	if resp.StatusCode != 200 || len(msgs) != 1 || msgs[0].Content != "antiga" {
		t.Fatalf("histórico: %d %+v", resp.StatusCode, msgs)
	}
}

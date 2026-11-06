package chat

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"sync"
	"time"
)

const saveTimeout = 3 * time.Second

// conn é uma conexão de um usuário em uma sala. Um usuário pode ter várias (abas).
// send é limitado: se o consumidor não acompanha, a conexão é derrubada em vez de
// travar a sala ou acumular memória sem limite.
type conn struct {
	userID, username, roomID string
	send                     chan []byte
	close                    func() // fecha o transporte; leitura falha e leave() é chamado
	once                     sync.Once
}

func newConn(roomID, userID, username string, buf int, closeFn func()) *conn {
	return &conn{roomID: roomID, userID: userID, username: username, send: make(chan []byte, buf), close: closeFn}
}

func (c *conn) kick() { c.once.Do(func() { c.close() }) }

type room struct {
	mu    sync.Mutex
	conns map[*conn]struct{}
	users map[string]int // conexões por usuário (presença só muda na 1ª/última)
}

type hub struct {
	st    store
	mu    sync.Mutex
	rooms map[string]*room
	now   func() time.Time
}

func newHub(st store) *hub {
	return &hub{st: st, rooms: map[string]*room{}, now: time.Now}
}

func (h *hub) join(c *conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	r := h.rooms[c.roomID]
	if r == nil {
		r = &room{conns: map[*conn]struct{}{}, users: map[string]int{}}
		h.rooms[c.roomID] = r
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.conns[c] = struct{}{}
	r.users[c.userID]++
	if r.users[c.userID] == 1 {
		r.fanout(h.presence(c, "entrou na sala"), nil)
	}
}

func (h *hub) leave(c *conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	r := h.rooms[c.roomID]
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.conns[c]; !ok {
		return
	}
	delete(r.conns, c)
	if r.users[c.userID]--; r.users[c.userID] <= 0 {
		delete(r.users, c.userID)
		r.fanout(h.presence(c, "saiu da sala"), nil)
	}
	if len(r.conns) == 0 {
		delete(h.rooms, c.roomID)
	}
}

func (h *hub) presence(c *conn, text string) []byte {
	return encode(Message{Type: TypePresence, Content: text, RoomID: c.roomID, UserID: c.userID,
		Username: c.username, CreatedAt: h.now().UTC()})
}

// handle processa um frame vindo de c. Identidade e sala vêm da conexão autenticada,
// nunca do payload: o cliente não consegue se passar por outro usuário.
func (h *hub) handle(c *conn, raw []byte) {
	var in Message
	if err := json.Unmarshal(raw, &in); err != nil {
		return
	}
	content := strings.TrimSpace(in.Content)
	if content == "PING" && in.Type == "" { // keepalive legado do frontend
		return
	}
	h.mu.Lock()
	r := h.rooms[c.roomID]
	h.mu.Unlock()
	if r == nil {
		return
	}
	out := Message{Type: in.Type, Content: in.Content, RoomID: c.roomID, UserID: c.userID, Username: c.username}

	switch in.Type {
	case "", TypeChat:
		if content == "" || len(in.Content) > maxContentBytes {
			c.reply(Message{Type: "ERROR", Content: "mensagem vazia ou grande demais", RoomID: c.roomID})
			return
		}
		out.Type = TypeChat
		// Segura a trava da sala durante gravar+distribuir: define a ordem e garante
		// que ninguém vê uma mensagem que não foi persistida.
		r.mu.Lock()
		defer r.mu.Unlock()
		ctx, cancel := context.WithTimeout(context.Background(), saveTimeout)
		defer cancel()
		if err := h.st.save(ctx, &out); err != nil {
			slog.Error("chat: falha ao salvar", "err", err, "room_id", c.roomID)
			c.reply(Message{Type: "ERROR", Content: "não foi possível enviar a mensagem", RoomID: c.roomID})
			return
		}
		r.fanout(encode(out), nil)
	case TypeTypingStart, TypeTypingStop:
		out.Content = ""
		out.CreatedAt = h.now().UTC()
		r.mu.Lock()
		defer r.mu.Unlock()
		r.fanout(encode(out), func(o *conn) bool { return o != c })
	case TypeDirect, TypeSignal:
		if in.TargetUserID == "" {
			return
		}
		out.TargetUserID = in.TargetUserID
		out.CreatedAt = h.now().UTC()
		r.mu.Lock()
		defer r.mu.Unlock()
		r.fanout(encode(out), func(o *conn) bool { return o.userID == in.TargetUserID || o.userID == c.userID })
	}
}

// fanout exige r.mu. Envio nunca bloqueia: consumidor lento é derrubado.
func (r *room) fanout(payload []byte, keep func(*conn) bool) {
	for c := range r.conns {
		if keep != nil && !keep(c) {
			continue
		}
		select {
		case c.send <- payload:
		default:
			go c.kick()
		}
	}
}

func (c *conn) reply(m Message) {
	select {
	case c.send <- encode(m):
	default:
	}
}

func encode(m Message) []byte {
	b, _ := json.Marshal(m)
	return b
}

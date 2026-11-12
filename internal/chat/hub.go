package chat

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

const (
	saveTimeout  = 3 * time.Second
	breakerPause = 5 * time.Second // após falha de gravação, vai direto ao spool por este tempo
	drainEvery   = 5 * time.Second
	outboxSize   = 1024
)

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

func (c *conn) reply(m Message) {
	select {
	case c.send <- encode(m):
	default:
	}
}

// filter escolhe quem recebe dentro da sala.
type filter struct {
	Except string
	Only   []string
}

func (f filter) match(c *conn) bool {
	if f.Except != "" && c.userID == f.Except {
		return false
	}
	return len(f.Only) == 0 || slices.Contains(f.Only, c.userID)
}

type room struct {
	mu    sync.Mutex
	conns map[*conn]struct{}
	users map[string]int // conexões por usuário (presença só muda na 1ª/última)
}

// fanout exige r.mu. Envio nunca bloqueia: consumidor lento é derrubado.
func (r *room) fanout(payload []byte, f filter) {
	for c := range r.conns {
		if !f.match(c) {
			continue
		}
		select {
		case c.send <- payload:
		default:
			go c.kick()
		}
	}
}

type hubOpts struct {
	bus    bus    // nil = processo único
	spool  *spool // nil = sem spool: falha de gravação vira erro para o remetente
	origin string // identifica esta réplica no barramento
}

type hub struct {
	st     store
	bus    bus
	spool  *spool
	origin string
	m      *metrics

	mu    sync.Mutex
	rooms map[string]*room
	now   func() time.Time
	// esperas entre tentativas de gravar (cobre queda curta de conexão/failover)
	retry []time.Duration
	// disjuntor: enquanto agora < skipUntil (unix nano) o banco é tratado como fora e o spool é usado direto,
	// para não segurar a sala em timeouts a cada mensagem durante uma queda.
	skipUntil atomic.Int64
	outbox    chan event
}

func newHub(st store, o ...hubOpts) *hub {
	var opts hubOpts
	if len(o) > 0 {
		opts = o[0]
	}
	if opts.origin == "" {
		opts.origin = uuid.NewString()
	}
	h := &hub{st: st, bus: opts.bus, spool: opts.spool, origin: opts.origin,
		rooms: map[string]*room{}, now: time.Now, outbox: make(chan event, outboxSize),
		retry: []time.Duration{100 * time.Millisecond, 300 * time.Millisecond, 900 * time.Millisecond}}
	h.m = newMetrics(h)
	return h
}

// run executa o trabalho em segundo plano até ctx terminar.
func (h *hub) run(ctx context.Context) {
	var wg sync.WaitGroup
	if h.bus != nil {
		wg.Add(2)
		go func() { defer wg.Done(); h.bus.run(ctx, h.onEvent) }()
		go func() { defer wg.Done(); h.publisher(ctx) }()
	}
	if h.spool != nil {
		wg.Add(1)
		go func() { defer wg.Done(); h.drainLoop(ctx) }()
	}
	wg.Wait()
	<-ctx.Done()
}

func (h *hub) publisher(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case ev := <-h.outbox:
			pctx, cancel := context.WithTimeout(ctx, saveTimeout)
			err := h.bus.publish(pctx, ev)
			cancel()
			if err != nil {
				h.m.busEvent("out", "error")
				slog.Warn("chat: falha ao publicar no barramento", "err", err)
				continue
			}
			h.m.busEvent("out", "ok")
		}
	}
}

// publish enfileira para as outras réplicas sem bloquear a sala. Fila cheia = descarta (e conta).
func (h *hub) publish(ev event) {
	if h.bus == nil {
		return
	}
	ev.Origin = h.origin
	select {
	case h.outbox <- ev:
	default:
		h.m.busEvent("out", "dropped")
	}
}

// onEvent entrega a clientes locais um evento vindo de outra réplica.
func (h *hub) onEvent(ev event) {
	if ev.Origin == h.origin {
		return
	}
	msg := ev.Msg
	if msg == nil && ev.MsgID != "" {
		ctx, cancel := context.WithTimeout(context.Background(), saveTimeout)
		m, err := h.st.byID(ctx, ev.MsgID)
		cancel()
		if err != nil {
			h.m.busEvent("in", "error")
			slog.Warn("chat: não achei a mensagem do evento", "id", ev.MsgID, "err", err)
			return
		}
		msg = &m
	}
	if msg == nil {
		return
	}
	h.mu.Lock()
	r := h.rooms[ev.Room]
	h.mu.Unlock()
	h.m.busEvent("in", "ok")
	if r == nil {
		return // ninguém desta sala conectado aqui
	}
	r.mu.Lock()
	r.fanout(encode(*msg), filter{Except: ev.Except, Only: ev.Only})
	r.mu.Unlock()
}

func (h *hub) drainLoop(ctx context.Context) {
	t := time.NewTicker(drainEvery)
	defer t.Stop()
	for {
		h.drainOnce(ctx)
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

func (h *hub) drainOnce(ctx context.Context) {
	n, err := h.spool.drain(ctx,
		func(c context.Context, m Message) error {
			sctx, cancel := context.WithTimeout(c, saveTimeout)
			defer cancel()
			return h.st.save(sctx, m)
		},
		func(c context.Context, m Message, reason string) error {
			sctx, cancel := context.WithTimeout(c, saveTimeout)
			defer cancel()
			h.m.message("dead_letter")
			return h.st.deadLetter(sctx, m, reason)
		})
	if n > 0 {
		slog.Info("chat: spool reprocessado", "mensagens", n, "restantes", h.spool.pending())
	}
	if err != nil && ctx.Err() == nil {
		slog.Warn("chat: spool ainda não pôde ser reprocessado", "err", err, "restantes", h.spool.pending())
	}
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
	h.m.conns.Add(context.Background(), 1)
	if r.users[c.userID] == 1 {
		h.announce(r, c, "entrou na sala")
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
	h.m.conns.Add(context.Background(), -1)
	if r.users[c.userID]--; r.users[c.userID] <= 0 {
		delete(r.users, c.userID)
		h.announce(r, c, "saiu da sala")
	}
	if len(r.conns) == 0 {
		delete(h.rooms, c.roomID)
	}
}

// announce exige r.mu.
func (h *hub) announce(r *room, c *conn, text string) {
	msg := Message{Type: TypePresence, Content: text, RoomID: c.roomID, UserID: c.userID,
		Username: c.username, CreatedAt: h.now().UTC()}
	r.fanout(encode(msg), filter{})
	h.publish(event{Room: c.roomID, Msg: &msg})
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
			h.m.message("rejected")
			c.reply(Message{Type: "ERROR", Content: "mensagem vazia ou grande demais", RoomID: c.roomID})
			return
		}
		out.Type = TypeChat
		id, err := newMessageID()
		if err != nil {
			c.reply(Message{Type: "ERROR", Content: "não foi possível enviar a mensagem", RoomID: c.roomID})
			return
		}
		out.ID, out.CreatedAt = id, h.now().UTC()
		// Segura a trava da sala durante gravar+distribuir: define a ordem e garante
		// que ninguém vê uma mensagem que não foi persistida (ou, na falta do banco, spoolada).
		r.mu.Lock()
		defer r.mu.Unlock()
		if !h.accept(c, &out) {
			return
		}
		r.fanout(encode(out), filter{})
		h.publish(event{Room: c.roomID, Msg: &out})
	case TypeTypingStart, TypeTypingStop:
		out.Content = ""
		out.CreatedAt = h.now().UTC()
		r.mu.Lock()
		defer r.mu.Unlock()
		r.fanout(encode(out), filter{Except: c.userID})
		h.publish(event{Room: c.roomID, Except: c.userID, Msg: &out})
	case TypeDirect, TypeSignal:
		if in.TargetUserID == "" {
			return
		}
		out.TargetUserID = in.TargetUserID
		out.CreatedAt = h.now().UTC()
		only := []string{in.TargetUserID, c.userID}
		r.mu.Lock()
		defer r.mu.Unlock()
		r.fanout(encode(out), filter{Only: only})
		h.publish(event{Room: c.roomID, Only: only, Msg: &out})
	}
}

// accept grava (ou, se o banco está fora, enfileira no spool). false = recusada; o remetente já foi avisado.
func (h *hub) accept(c *conn, m *Message) bool {
	err := h.persist(m)
	switch {
	case err == nil:
		h.m.message("ok")
		return true
	case isPermanent(err):
		h.m.message("rejected")
		slog.Error("chat: mensagem recusada pelo banco", "err", err, "room_id", m.RoomID)
		c.reply(Message{Type: "ERROR", Content: "a mensagem não pôde ser salva (a sala ainda existe?)", RoomID: m.RoomID})
		return false
	case h.spool != nil:
		if serr := h.spool.append(*m); serr != nil {
			h.m.message("failed")
			slog.Error("chat: spool falhou", "err", serr, "db_err", err, "room_id", m.RoomID)
			c.reply(Message{Type: "ERROR", Content: "não foi possível enviar a mensagem", RoomID: m.RoomID})
			return false
		}
		h.m.message("spooled")
		slog.Warn("chat: banco fora, mensagem no spool", "err", err, "id", m.ID)
		return true
	default:
		h.m.message("failed")
		slog.Error("chat: falha ao salvar", "err", err, "room_id", m.RoomID)
		c.reply(Message{Type: "ERROR", Content: "não foi possível enviar a mensagem", RoomID: m.RoomID})
		return false
	}
}

var errBreakerOpen = errors.New("banco fora (disjuntor aberto)")

// persist grava com backoff curto. Só chamado com a trava da sala, então o total é limitado
// (tentativas x saveTimeout) e a ordem das mensagens é preservada. Com o disjuntor aberto
// (e spool disponível) nem tenta o banco.
func (h *hub) persist(m *Message) error {
	if h.spool != nil && h.now().UnixNano() < h.skipUntil.Load() {
		return errBreakerOpen
	}
	var err error
	for i := 0; ; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), saveTimeout)
		err = h.st.save(ctx, *m)
		cancel()
		if err == nil {
			h.skipUntil.Store(0)
			return nil
		}
		if isPermanent(err) || i >= len(h.retry) {
			break
		}
		slog.Warn("chat: falha ao salvar, tentando de novo", "tentativa", i+1, "err", err)
		time.Sleep(h.retry[i])
	}
	if !isPermanent(err) {
		h.skipUntil.Store(h.now().Add(breakerPause).UnixNano())
	}
	return err
}

func encode(m Message) []byte {
	b, _ := json.Marshal(m)
	return b
}

package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func msg(room, content string) Message {
	id, _ := newMessageID()
	return Message{ID: id, Type: TypeChat, RoomID: room, UserID: "u", Username: "u", Content: content, CreatedAt: time.Now().UTC()}
}

func okSave(got *[]string) func(context.Context, Message) error {
	return func(_ context.Context, m Message) error { *got = append(*got, m.Content); return nil }
}

func noDead(context.Context, Message, string) error { return nil }

func TestSpoolKeepsOrderAndStopsAtTransientError(t *testing.T) {
	sp, err := newSpool(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []string{"a", "b", "c"} {
		if err := sp.append(msg("r", c)); err != nil {
			t.Fatal(err)
		}
	}
	var saved []string
	n, err := sp.drain(context.Background(), func(_ context.Context, m Message) error {
		if m.Content == "b" {
			return errors.New("banco fora")
		}
		saved = append(saved, m.Content)
		return nil
	}, noDead)
	if err == nil || n != 1 || sp.pending() != 2 {
		t.Fatalf("esperava parar em b: n=%d err=%v pend=%d", n, err, sp.pending())
	}
	saved = nil
	if n, err = sp.drain(context.Background(), okSave(&saved), noDead); err != nil || n != 2 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	if strings.Join(saved, "") != "bc" || sp.pending() != 0 {
		t.Fatalf("ordem/restante errados: %v pend=%d", saved, sp.pending())
	}
}

func TestSpoolPermanentErrorGoesToDeadLetter(t *testing.T) {
	sp, _ := newSpool(t.TempDir())
	_ = sp.append(msg("r", "ruim"))
	_ = sp.append(msg("r", "boa"))
	var dead, saved []string
	_, err := sp.drain(context.Background(), func(_ context.Context, m Message) error {
		if m.Content == "ruim" {
			return fmt.Errorf("fk: %w", &pgconn.PgError{Code: "23503"})
		}
		saved = append(saved, m.Content)
		return nil
	}, func(_ context.Context, m Message, _ string) error { dead = append(dead, m.Content); return nil })
	if err != nil || len(dead) != 1 || dead[0] != "ruim" || len(saved) != 1 || sp.pending() != 0 {
		t.Fatalf("err=%v dead=%v saved=%v pend=%d", err, dead, saved, sp.pending())
	}
}

func TestSpoolSurvivesRestartAndKeepsAppendsMadeDuringDrain(t *testing.T) {
	dir := t.TempDir()
	sp, _ := newSpool(dir)
	_ = sp.append(msg("r", "antes-do-restart"))

	sp2, err := newSpool(dir) // "restart"
	if err != nil || sp2.pending() != 1 {
		t.Fatalf("pendente após restart: %d err=%v", sp2.pending(), err)
	}
	var saved []string
	n, err := sp2.drain(context.Background(), func(_ context.Context, m Message) error {
		saved = append(saved, m.Content)
		if m.Content == "antes-do-restart" {
			_ = sp2.append(msg("r", "chegou-durante")) // append concorrente ao drain
		}
		return nil
	}, noDead)
	if err != nil || n != 1 || sp2.pending() != 1 {
		t.Fatalf("n=%d err=%v pend=%d", n, err, sp2.pending())
	}
	saved = nil
	if n, _ = sp2.drain(context.Background(), okSave(&saved), noDead); n != 1 || saved[0] != "chegou-durante" {
		t.Fatalf("append durante o drain foi perdido: %v", saved)
	}
}

func spoolHub(t *testing.T, st *fakeStore) (*hub, *spool) {
	t.Helper()
	sp, err := newSpool(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	h := newHub(st, hubOpts{spool: sp})
	h.retry = []time.Duration{time.Millisecond}
	return h, sp
}

func recv(t *testing.T, c *conn) Message {
	t.Helper()
	select {
	case b := <-c.send:
		var m Message
		if err := json.Unmarshal(b, &m); err != nil {
			t.Fatal(err)
		}
		return m
	case <-time.After(2 * time.Second):
		t.Fatal("nada recebido")
		return Message{}
	}
}

func recvType(t *testing.T, c *conn, typ string) Message {
	t.Helper()
	for i := 0; i < 10; i++ {
		if m := recv(t, c); m.Type == typ {
			return m
		}
	}
	t.Fatalf("não recebi %s", typ)
	return Message{}
}

func TestDBDownMessageIsDeliveredSpooledAndReplayedOnce(t *testing.T) {
	st := &fakeStore{fail: true}
	h, sp := spoolHub(t, st)
	room := uuid.NewString()
	a := newConn(room, "ana", "ana", 32, func() {})
	b := newConn(room, "bia", "bia", 32, func() {})
	h.join(a)
	h.join(b)
	recvType(t, a, TypePresence)

	h.handle(a, []byte(`{"type":"CHAT","content":"1"}`))
	if m := recvType(t, b, TypeChat); m.Content != "1" || m.ID == "" {
		t.Fatalf("bia deveria receber mesmo com o banco fora: %+v", m)
	}
	// disjuntor aberto: a 2ª não espera retries/timeouts
	start := time.Now()
	h.handle(a, []byte(`{"type":"CHAT","content":"2"}`))
	if time.Since(start) > 200*time.Millisecond {
		t.Fatalf("2ª mensagem demorou %v: disjuntor não abriu", time.Since(start))
	}
	if st.count() != 0 || sp.pending() != 2 {
		t.Fatalf("db=%d spool=%d", st.count(), sp.pending())
	}

	st.setFail(false)
	h.drainOnce(context.Background())
	h.drainOnce(context.Background()) // reprocessar de novo não duplica
	if st.count() != 2 || sp.pending() != 0 {
		t.Fatalf("após replay: db=%d spool=%d", st.count(), sp.pending())
	}
	if st.msgs[0].Content != "1" || st.msgs[1].Content != "2" {
		t.Fatalf("ordem: %+v", st.msgs)
	}
}

func TestPermanentDBErrorIsRejectedNotSpooled(t *testing.T) {
	st := &fakeStore{perm: &pgconn.PgError{Code: "23503"}}
	h, sp := spoolHub(t, st)
	a := newConn(uuid.NewString(), "ana", "ana", 32, func() {})
	h.join(a)
	recvType(t, a, TypePresence)
	h.handle(a, []byte(`{"type":"CHAT","content":"x"}`))
	if m := recvType(t, a, "ERROR"); m.Content == "" {
		t.Fatal("remetente deve ser avisado")
	}
	if sp.pending() != 0 {
		t.Fatal("erro permanente não deve ir ao spool")
	}
}

// ---- várias réplicas ----

type memBus struct {
	mu   sync.Mutex
	subs []func(event)
}

func (b *memBus) publish(_ context.Context, ev event) error {
	payload, err := encodeEvent(ev) // passa pela mesma codificação do NOTIFY
	if err != nil {
		return err
	}
	var back event
	if err := json.Unmarshal(payload, &back); err != nil {
		return err
	}
	b.mu.Lock()
	subs := append([]func(event){}, b.subs...)
	b.mu.Unlock()
	for _, s := range subs {
		s(back)
	}
	return nil
}

func (b *memBus) run(ctx context.Context, handle func(event)) {
	b.mu.Lock()
	b.subs = append(b.subs, handle)
	b.mu.Unlock()
	<-ctx.Done()
}

func (b *memBus) waitSubs(t *testing.T, n int) {
	t.Helper()
	for i := 0; i < 200; i++ {
		b.mu.Lock()
		got := len(b.subs)
		b.mu.Unlock()
		if got == n {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("réplicas não se inscreveram no barramento")
}

func twoReplicas(t *testing.T) (h1, h2 *hub, st *fakeStore) {
	t.Helper()
	st = &fakeStore{}
	b := &memBus{}
	h1 = newHub(st, hubOpts{bus: b, origin: "pod-1"})
	h2 = newHub(st, hubOpts{bus: b, origin: "pod-2"})
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go h1.run(ctx)
	go h2.run(ctx)
	b.waitSubs(t, 2)
	return h1, h2, st
}

func TestChatCrossesReplicas(t *testing.T) {
	h1, h2, st := twoReplicas(t)
	room := uuid.NewString()
	ana := newConn(room, "ana", "ana", 32, func() {})
	bia := newConn(room, "bia", "bia", 32, func() {})
	h1.join(ana)
	h2.join(bia)

	h1.handle(ana, []byte(`{"type":"CHAT","content":"oi da réplica 1"}`))
	if m := recvType(t, bia, TypeChat); m.Content != "oi da réplica 1" || m.UserID != "ana" {
		t.Fatalf("bia (outra réplica) deveria receber: %+v", m)
	}
	if m := recvType(t, ana, TypeChat); m.Content != "oi da réplica 1" {
		t.Fatalf("ana: %+v", m)
	}
	select { // ana não pode receber de volta pelo barramento (eco)
	case b := <-ana.send:
		t.Fatalf("eco inesperado: %s", b)
	case <-time.After(150 * time.Millisecond):
	}
	if st.count() != 1 {
		t.Fatalf("só a réplica de origem grava: %d", st.count())
	}
}

func TestPresenceTypingAndDirectCrossReplicas(t *testing.T) {
	h1, h2, _ := twoReplicas(t)
	room := uuid.NewString()
	ana := newConn(room, "ana", "ana", 32, func() {})
	h1.join(ana)
	bia := newConn(room, "bia", "bia", 32, func() {})
	cai := newConn(room, "cai", "cai", 32, func() {})
	h2.join(bia)
	h2.join(cai)

	if m := recvType(t, ana, TypePresence); m.UserID != "ana" { // própria entrada
		t.Fatalf("%+v", m)
	}
	if m := recvType(t, ana, TypePresence); m.UserID != "bia" || m.Content != "entrou na sala" {
		t.Fatalf("ana deveria ver bia entrar pela outra réplica: %+v", m)
	}

	h1.handle(ana, []byte(`{"type":"TYPING_START"}`))
	if m := recvType(t, bia, TypeTypingStart); m.UserID != "ana" {
		t.Fatalf("%+v", m)
	}

	h1.handle(ana, []byte(`{"type":"SIGNAL","target_user_id":"bia","content":"sdp"}`))
	if m := recvType(t, bia, TypeSignal); m.Content != "sdp" {
		t.Fatalf("%+v", m)
	}
	for { // cai não é o alvo: não pode receber o sinal
		select {
		case b := <-cai.send:
			if strings.Contains(string(b), `"SIGNAL"`) {
				t.Fatalf("sinal vazou para quem não era o alvo: %s", b)
			}
		case <-time.After(150 * time.Millisecond):
			return
		}
	}
}

func TestLargeChatMessageTravelsAsIDAndIsFetched(t *testing.T) {
	h1, h2, _ := twoReplicas(t)
	room := uuid.NewString()
	ana := newConn(room, "ana", "ana", 32, func() {})
	bia := newConn(room, "bia", "bia", 32, func() {})
	h1.join(ana)
	h2.join(bia)

	// '<' vira < no JSON (6 bytes): passa fácil de 7500 bytes no NOTIFY
	big := strings.Repeat("<", 3000)
	raw, _ := json.Marshal(map[string]string{"type": "CHAT", "content": big})
	h1.handle(ana, raw)
	if m := recvType(t, bia, TypeChat); m.Content != big {
		t.Fatalf("conteúdo grande deveria chegar via busca por id (len=%d)", len(m.Content))
	}
}

func TestEncodeEventLimits(t *testing.T) {
	small := event{Origin: "o", Room: "r", Msg: &Message{Type: TypeChat, ID: "i", Content: "x"}}
	if b, err := encodeEvent(small); err != nil || len(b) > maxNotifyPayload {
		t.Fatalf("pequeno: %v", err)
	}
	bigSignal := event{Origin: "o", Room: "r", Msg: &Message{Type: TypeSignal, Content: strings.Repeat("a", 9000)}}
	if _, err := encodeEvent(bigSignal); !errors.Is(err, errPayloadTooLarge) {
		t.Fatalf("sinal enorme deve ser recusado, veio %v", err)
	}
	bigChat := event{Origin: "o", Room: "r", Msg: &Message{Type: TypeChat, ID: "abc", Content: strings.Repeat("a", 9000)}}
	b, err := encodeEvent(bigChat)
	if err != nil || len(b) > maxNotifyPayload || !strings.Contains(string(b), `"id":"abc"`) {
		t.Fatalf("chat enorme deve virar só o id: %s %v", b, err)
	}
}

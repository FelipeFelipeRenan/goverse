package chat

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/FelipeFelipeRenan/goverse/internal/platform/authn"
	"github.com/FelipeFelipeRenan/goverse/internal/platform/httpx"
	"github.com/FelipeFelipeRenan/goverse/internal/rooms"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	sendBuffer = 64
	readLimit  = 16 << 10
	pongWait   = 60 * time.Second
	pingEvery  = 30 * time.Second
	writeWait  = 10 * time.Second
)

type api struct {
	hub     *hub
	st      store
	keys    *authn.Keys
	members rooms.Membership
	origins map[string]struct{}
}

func (a *api) routes(mux *http.ServeMux) {
	mux.Handle("GET /ws", a.keys.Require(http.HandlerFunc(a.serveWS)))
	mux.Handle("GET /rooms/{id}/messages", a.keys.Require(http.HandlerFunc(a.history)))
}

// authorize valida o roomID e a participação. Responde e devolve false em caso de falha.
func (a *api) authorize(w http.ResponseWriter, r *http.Request, roomID string) (authn.Identity, bool) {
	id, _ := authn.From(r.Context()) // Require garante que existe
	if _, err := uuid.Parse(roomID); err != nil {
		httpx.Error(w, http.StatusBadRequest, "id de sala inválido")
		return id, false
	}
	ok, err := a.members.IsMember(r.Context(), roomID, id.UserID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "erro ao verificar permissão")
		return id, false
	}
	if !ok {
		httpx.Error(w, http.StatusForbidden, "você não é membro desta sala")
		return id, false
	}
	return id, true
}

func (a *api) history(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("id")
	if _, ok := a.authorize(w, r, roomID); !ok {
		return
	}
	limit, offset := 50, 0
	if v, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && v > 0 {
		limit = min(v, 200)
	}
	if v, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil && v > 0 {
		offset = v
	}
	msgs, err := a.st.history(r.Context(), roomID, limit, offset)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "erro ao buscar mensagens")
		return
	}
	httpx.JSON(w, http.StatusOK, msgs)
}

func (a *api) checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" { // cliente não-browser; a sessão (cookie) continua obrigatória
		return true
	}
	if _, ok := a.origins[origin]; ok {
		return true
	}
	u, err := url.Parse(origin)
	return err == nil && u.Host == r.Host
}

func (a *api) serveWS(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("roomId")
	id, ok := a.authorize(w, r, roomID)
	if !ok {
		return
	}
	up := websocket.Upgrader{CheckOrigin: a.checkOrigin, ReadBufferSize: 4096, WriteBufferSize: 4096}
	ws, err := up.Upgrade(w, r, nil) // já responde com erro se falhar
	if err != nil {
		return
	}
	defer ws.Close()

	c := newConn(roomID, id.UserID, id.Username, sendBuffer, func() { _ = ws.Close() })
	a.hub.join(c)
	defer a.hub.leave(c)

	done, stop := make(chan struct{}), make(chan struct{})
	go func() { // escrita: única goroutine que escreve no socket
		defer close(done)
		t := time.NewTicker(pingEvery)
		defer t.Stop()
		for {
			select {
			case msg := <-c.send:
				_ = ws.SetWriteDeadline(time.Now().Add(writeWait))
				if ws.WriteMessage(websocket.TextMessage, msg) != nil {
					c.kick()
					return
				}
			case <-t.C:
				_ = ws.SetWriteDeadline(time.Now().Add(writeWait))
				if ws.WriteMessage(websocket.PingMessage, nil) != nil {
					c.kick()
					return
				}
			case <-stop:
				return
			}
		}
	}()

	ws.SetReadLimit(readLimit)
	_ = ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error { return ws.SetReadDeadline(time.Now().Add(pongWait)) })
	for {
		_, raw, err := ws.ReadMessage()
		if err != nil {
			break
		}
		_ = ws.SetReadDeadline(time.Now().Add(pongWait))
		a.hub.handle(c, raw)
	}
	close(stop)
	c.kick()
	<-done
}

package websocket

import (
	"net/http"

	"github.com/FelipeFelipeRenan/goverse/chat-service/internal/auth"
	"github.com/FelipeFelipeRenan/goverse/chat-service/internal/hub"
	"github.com/FelipeFelipeRenan/goverse/chat-service/pkg/logger"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketHandler struct {
	hub *hub.Hub
}

func NewWebSocketHandler(hub *hub.Hub) *WebSocketHandler {
	return &WebSocketHandler{hub: hub}
}

// necessario nomear como hd para nao dar conflito
func (hd *WebSocketHandler) ServeWs(w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie("access_token")
	if err != nil {
		logger.Error("Erro ao obter cookie de acesso", "err", err)
		http.Error(w, "Cookie de autenticação não encontrado", http.StatusUnauthorized)
		return
	}

	tokenString := cookie.Value
	claims, err := auth.ValidateToken(tokenString)
	if err != nil {
		logger.Error("Erro na validação do token", "err", err)
		http.Error(w, "token inválido", http.StatusUnauthorized)
		return
	}

	roomID := r.URL.Query().Get("roomId")
	if roomID == "" {
		http.Error(w, "O parâmetro 'roomId' é obrigatório", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("Erro no upgrade da conexão para websocket", "err", err)
		return
	}

	client := &hub.Client{
		Conn:     conn,
		Hub:      hd.hub,
		RoomID:   roomID,
		UserID:   claims.UserID,
		Username: claims.UserName,
		Send:     make(chan []byte, 256),
	}
	client.Hub.Register <- client

	logger.Info("Cliente AUTENTICADO conectado", "username", claims.UserName, "userID", client.UserID, "roomID", client.RoomID)

	// Inicia os processos de leitura e escrita em goroutines separadas.
	go client.WritePump()
	go client.ReadPump()
}

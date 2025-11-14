package hub

import (
	"context"
	"time"

	"github.com/FelipeFelipeRenan/goverse/chat-service/internal/message/domain"
	"github.com/FelipeFelipeRenan/goverse/chat-service/pkg/logger"
	"github.com/gorilla/websocket"
)

// Client é a representação de um usuário conectado via WebSocket.
type Client struct {
	Conn     *websocket.Conn
	Hub      *Hub
	RoomID   string
	UserID   string
	Username string
	Send     chan []byte
}

// ReadPump lê mensagens do cliente e as envia para o hub.
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, payload, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("erro inesperado no websocket", "err", err)

			}
			break
		}

		var msg domain.Message
		if err := domain.FromJSON(payload, &msg); err != nil {
			logger.Error("erro ao decodificar json", "err", err)
			continue
		}

		if msg.Content == "PING" {
			c.Hub.redisClient.Expire(context.Background(), "user:status:"+c.UserID, 5*time.Minute)
			continue
		}

		msg.UserID = c.UserID
		msg.Username = c.Username
		msg.RoomID = c.RoomID

		if msg.Type == "TYPING_START" || msg.Type == "TYPING_STOP" {
			c.Hub.Relay <- &msg
		} else {
			if msg.Type == "" {
				msg.Type = "CHAT"
			}
			// Por enquanto, apenas enviamos a mensagem bruta para o hub processar.
			c.Hub.Broadcast <- &msg
		}
	}
}

// WritePump envia mensagens do hub para o cliente.
func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		message, ok := <-c.Send
		if !ok {
			// O hub fechou este canal.
			c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		err := c.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			logger.Error("erro ao escrever mensagem ao websocket", "err", err)
			return
		}
	}
}

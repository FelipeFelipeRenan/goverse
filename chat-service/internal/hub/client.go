package hub

import (
	"github.com/gorilla/websocket"
	"log"
)

// Client é a representação de um usuário conectado via WebSocket.
type Client struct {
	Conn   *websocket.Conn
	Hub    *Hub
	RoomID string
	UserID string
	Send   chan []byte
}

// ReadPump lê mensagens do cliente e as envia para o hub.
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("erro inesperado de websocket: %v", err)
			}
			break
		}
		// Por enquanto, apenas enviamos a mensagem bruta para o hub processar.
		c.Hub.Broadcast <- message
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
			log.Printf("erro ao escrever mensagem para o websocket: %v", err)
			return
		}
	}
}

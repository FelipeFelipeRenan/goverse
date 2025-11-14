package hub

import (
	"context"
	"strings"
	"time"

	"github.com/FelipeFelipeRenan/goverse/chat-service/internal/message/domain"
	"github.com/FelipeFelipeRenan/goverse/chat-service/internal/message/service"
	"github.com/FelipeFelipeRenan/goverse/chat-service/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// Hub mantém o conjunto de clientes ativos e transmite mensagens para eles.
type Hub struct {
	// Mapeia um ID de sala para um map de clientes naquela sala
	Rooms map[string]map[*Client]bool

	// Mensagens de entrada dos clientes
	Broadcast chan *domain.Message

	// Mensagens efemeras (nao serao salvas )
	Relay chan *domain.Message
	// Solicitação de registros de clientes
	Register chan *Client

	// Solicitação de cancelamento de registro de clientes
	Unregister chan *Client

	Svc service.MessageService

	redisClient *redis.Client
}

func NewHub(svc service.MessageService, redisClient *redis.Client) *Hub {
	return &Hub{
		Broadcast:   make(chan *domain.Message),
		Relay:       make(chan *domain.Message),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		Rooms:       make(map[string]map[*Client]bool),
		Svc:         svc,
		redisClient: redisClient,
	}
}

func (h *Hub) runRedisSubscriber() {
	ctx := context.Background()

	pubsub := h.redisClient.PSubscribe(ctx, "chat:room:*")
	defer pubsub.Close()

	ch := pubsub.Channel()

	logger.Info("Hub está ouvindo os canais do Redis")

	for msg := range ch {

		// msg.Channel é "chat:room:sala-1", por exemplo
		roomID := strings.TrimPrefix(msg.Channel, "chat:room:")

		// Encontra a sala e envia a mensagem para todos os clientes locais nela
		if room, ok := h.Rooms[roomID]; ok {
			for client := range room {
				select {
				case client.Send <- []byte(msg.Payload):
				default:
					// se o canal do cliente estiver cheio, assume que ele está morto
					close(client.Send)
					delete(h.Rooms[roomID], client)
				}
			}
		}
	}
}

func (h *Hub) Run() {

	// inicia o ouvinte redis em uma goroutine separada
	go h.runRedisSubscriber()
	ctx := context.Background()

	for {
		select {
		case client := <-h.Register:
			go func(c *Client) {
				// Se a sala não existir, crie-a.
				if _, ok := h.Rooms[c.RoomID]; !ok {
					h.Rooms[c.RoomID] = make(map[*Client]bool)
				}
				// Registra o cliente na sala.
				h.Rooms[c.RoomID][c] = true

				//  Define o status no Redis com expiração (ex: 5 minutos)
				//    Isso garante que, se o serviço cair, o usuário ficará "offline"
				h.redisClient.SetEx(ctx, "user:status:"+c.UserID, "online", 5*time.Minute)

				// Cria uma mensagem de presença para modificar a sala
				presenceMsg := &domain.Message{
					Content:  "entrou na sala", // tratar com o front depois
					RoomID:   c.RoomID,
					UserID:   c.UserID,
					Username: c.Username,
					Type:     "PRESENCE",
					// adicionar type depois, para diferenciar o tipo de mensagem
				}

				h.Broadcast <- presenceMsg
				logger.Info("Cliente registrado e presença 'online' definida", "userID", c.UserID)
			}(client)

		case client := <-h.Unregister:
			go func(c *Client) {

				// Remove o cliente da sala.
				if room, ok := h.Rooms[c.RoomID]; ok {
					if _, ok := room[c]; ok {
						delete(h.Rooms[client.RoomID], c)
						close(c.Send)
					}
					// Se a sala ficar vazia, opcionalmente, remova a sala do map.
					if len(h.Rooms[c.RoomID]) == 0 {
						delete(h.Rooms, c.RoomID)
					}
				}

				//    Define o status no Redis como "offline"
				//    Usamos SetEX com 24h só para manter o dado, poderia ser um DEL ou SET simples
				h.redisClient.SetEx(ctx, "user:status:"+c.UserID, "offline", 24*time.Hour)

				// Cria uma mensagem de presença para modificar a sala
				presenceMsg := &domain.Message{
					Content:  "saiu da sala", // tratar com o front depois
					RoomID:   c.RoomID,
					UserID:   c.UserID,
					Username: c.Username,
					Type:     "PRESENCE",
					// adicionar type depois, para diferenciar o tipo de mensagem
				}

				h.Broadcast <- presenceMsg
				logger.Info("Cliente registrado e presença 'online' definida", "userID", c.UserID)

			}(client)
		case message := <-h.Broadcast:
			ctx := context.Background()
			if message.Type == "CHAT" {
				// 2. Agora, tentar salvar no banco
				if err := h.Svc.ProcessAndSaveMessage(ctx, message); err != nil {
					logger.Error("!!! ERRO AO SALVAR MENSAGEM !!!", "err", err) // Mudei o log para ficar óbvio
					// Não damos 'continue' para que o Redis ainda funcione
				}
			}
			payload, err := message.ToJSON()
			if err != nil {
				logger.Error("erro ao parsear mensagem", "err", err)
				continue
			}
			// 3. Publica a mensagem no Redis
			channelName := "chat:room:" + message.RoomID
			if err := h.redisClient.Publish(ctx, channelName, payload).Err(); err != nil {
				logger.Error("erro ao publicar mensagem ao Redis", "err", err)
			}
		case message := <-h.Relay:
			// mensagem efemera, apenas retransmitir no redis
			payload, err := message.ToJSON()
			if err != nil {
				logger.Error("erro ao parsear mensagem de relay", "err", err)
				continue
			}
			channelName := "chat:room:" + message.RoomID
			if err := h.redisClient.Publish(ctx, channelName, payload).Err(); err != nil {
				logger.Error("erro ao publicar relay ao Redis", "err", err)
			}
		}
	}
}

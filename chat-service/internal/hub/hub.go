package hub

import (
	"context"
	"encoding/json"
	"time"

	"github.com/FelipeFelipeRenan/goverse/chat-service/internal/message/domain"
	pkgkafka "github.com/FelipeFelipeRenan/goverse/chat-service/pkg/kafka"
	"github.com/FelipeFelipeRenan/goverse/common/pkg/logger"
	"github.com/redis/go-redis/v9"
)

type Hub struct {
	Rooms   map[string]map[*Client]bool
	Clients map[string]*Client

	Broadcast  chan *domain.Message
	Relay      chan *domain.Message
	Register   chan *Client
	Unregister chan *Client

	// [NOVO] Canal para receber mensagens vindas do Redis de forma segura
	redisIngress chan *domain.Message

	KafkaProducer *pkgkafka.Producer
	redisClient   *redis.Client
}

func NewHub(redisClient *redis.Client, kafkaProducer *pkgkafka.Producer) *Hub {
	return &Hub{
		Broadcast:  make(chan *domain.Message),
		Relay:      make(chan *domain.Message),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		// Canal sem buffer ou com buffer pequeno é ok aqui
		redisIngress:  make(chan *domain.Message, 256),
		Rooms:         make(map[string]map[*Client]bool),
		Clients:       make(map[string]*Client),
		redisClient:   redisClient,
		KafkaProducer: kafkaProducer,
	}
}

func (h *Hub) runRedisSubscriber() {
	ctx := context.Background()
	pubsub := h.redisClient.PSubscribe(ctx, "chat:room:*")
	defer pubsub.Close()

	ch := pubsub.Channel()
	logger.Info("Hub está ouvindo os canais do Redis")

	for msg := range ch {
		var message domain.Message
		if err := json.Unmarshal([]byte(msg.Payload), &message); err != nil {
			logger.Error("Erro ao decodificar mensagem do Redis", "err", err)
			continue
		}
		// [CORREÇÃO] Não acessamos mapas aqui. Apenas encaminhamos para o loop principal.
		h.redisIngress <- &message
	}
}

func (h *Hub) Run() {
	go h.runRedisSubscriber()
	ctx := context.Background()

	for {
		select {
		case client := <-h.Register:
			// 1. Registra nos Mapas (Thread-safe aqui dentro)
			if _, ok := h.Rooms[client.RoomID]; !ok {
				h.Rooms[client.RoomID] = make(map[*Client]bool)
			}
			h.Rooms[client.RoomID][client] = true
			h.Clients[client.UserID] = client

			// 2. Redis
			h.redisClient.SetEx(ctx, "user:status:"+client.UserID, "online", 5*time.Minute)
			h.redisClient.SAdd(ctx, "room:active:"+client.RoomID, client.UserID)

			// 3. Notifica entrada
			presenceMsg := &domain.Message{
				Type:     "PRESENCE",
				Content:  "entrou na sala",
				RoomID:   client.RoomID,
				UserID:   client.UserID,
				Username: client.Username,
			}

			// [CORREÇÃO DO DEADLOCK]
			// Usamos uma goroutine para não bloquear o loop atual tentando escrever no canal que ele mesmo lê.
			go func() {
				h.Broadcast <- presenceMsg
			}()

			logger.Info("Cliente registrado", "userID", client.UserID)

		case client := <-h.Unregister:
			if room, ok := h.Rooms[client.RoomID]; ok {
				if _, ok := room[client]; ok {
					delete(room, client)
					close(client.Send)
				}
				if len(room) == 0 {
					delete(h.Rooms, client.RoomID)
				}
			}
			delete(h.Clients, client.UserID)

			h.redisClient.SetEx(ctx, "user:status:"+client.UserID, "offline", 24*time.Hour)
			h.redisClient.SRem(ctx, "room:active:"+client.RoomID, client.UserID)

			presenceMsg := &domain.Message{
				Type:     "PRESENCE",
				Content:  "saiu da sala",
				RoomID:   client.RoomID,
				UserID:   client.UserID,
				Username: client.Username,
			}

			// [CORREÇÃO DO DEADLOCK]
			go func() {
				h.Broadcast <- presenceMsg
			}()

			logger.Info("Cliente desconectado", "userID", client.UserID)

		case message := <-h.Broadcast:
			// Lógica de envio (apenas publica, não entrega)
			shouldPersist := message.Type == "CHAT" || message.Type == "DIRECT"
			if shouldPersist {
				jsonBytes, _ := message.ToJSON()
				err := h.KafkaProducer.WriteMessage(ctx, []byte(message.RoomID), jsonBytes)
				if err != nil {
					logger.Error("Erro ao enviar mensagem para o Kafka", "err", err)
				}
			}

			payload, err := message.ToJSON()
			if err == nil {
				// Publica para todos (inclusive eu mesmo recebo de volta via Redis)
				h.redisClient.Publish(ctx, "chat:room:"+message.RoomID, payload)
			}

		case message := <-h.Relay:
			payload, _ := message.ToJSON()
			h.redisClient.Publish(ctx, "chat:room:"+message.RoomID, payload)

		// [NOVO] Lógica de entrega (Vem do Redis)
		case message := <-h.redisIngress:
			// Aqui temos acesso seguro aos mapas h.Clients e h.Rooms
			payload, _ := message.ToJSON()

			if message.TargetUserID != "" {
				// Entrega Direta (Direct/Signal)
				if client, ok := h.Clients[message.TargetUserID]; ok {
					select {
					case client.Send <- payload:
					default:
						close(client.Send)
						delete(h.Clients, message.TargetUserID)
					}
				}
			} else {
				// Broadcast para a Sala
				if room, ok := h.Rooms[message.RoomID]; ok {
					for client := range room {
						select {
						case client.Send <- payload:
						default:
							close(client.Send)
							delete(h.Rooms[message.RoomID], client)
						}
					}
				}
			}
		}
	}
}

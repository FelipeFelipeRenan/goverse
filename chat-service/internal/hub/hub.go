package hub

import "log"

// Hub mantém o conjunto de clientes ativos e transmite mensagens para eles.
type Hub struct {
	// Mapeia um ID de sala para um map de clientes naquela sala
	Rooms map[string]map[*Client]bool

	// Mensagens de entrada dos clientes
	Broadcast chan *Message

	// Solicitação de registros de clientes
	Register chan *Client

	// Solicitação de cancelamento de registro de clientes
	Unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:  make(chan *Message),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Rooms:      make(map[string]map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			// Se a sala não existir, crie-a.
			if _, ok := h.Rooms[client.RoomID]; !ok {
				h.Rooms[client.RoomID] = make(map[*Client]bool)
			}
			// Registra o cliente na sala.
			h.Rooms[client.RoomID][client] = true

		case client := <-h.Unregister:
			// Remove o cliente da sala.
			if room, ok := h.Rooms[client.RoomID]; ok {
				if _, ok := room[client]; ok {
					delete(h.Rooms[client.RoomID], client)
					close(client.Send)
				}
				// Se a sala ficar vazia, opcionalmente, remova a sala do map.
				if len(h.Rooms[client.RoomID]) == 0 {
					delete(h.Rooms, client.RoomID)
				}
			}

		case message := <-h.Broadcast:
			// Este é um broadcast simplificado. A lógica real seria encontrar
			// a sala do cliente que enviou a mensagem e enviar para todos
			// os outros clientes APENAS naquela sala.
			if room, ok := h.Rooms[message.RoomID]; ok{
				payload, err := message.ToJSON()
				if err != nil {
					log.Printf("erro ao parsear mesagem: %v", err)
					continue
				}

				for client := range room{
					select{
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

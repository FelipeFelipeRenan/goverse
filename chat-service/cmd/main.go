package main

import (
	"github.com/FelipeFelipeRenan/goverse/chat-service/internal/hub"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// IMPORTANTE: Em produção, valide a origem para permitir apenas seu frontend.
		// Ex: return r.Header.Get("Origin") == "http://localhost:4200"
		return true
	},
}

func serveWs(h *hub.Hub, w http.ResponseWriter, r *http.Request) {
	// TODO: Implementar a validação do token JWT aqui.
	userID := "usuario_simulado_123"
	roomID := r.URL.Query().Get("roomId")
	if roomID == "" {
		http.Error(w, "O parâmetro 'roomId' é obrigatório", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Erro no upgrade da conexão para websocket: %v", err)
		return
	}

	client := &hub.Client{
		Conn:   conn,
		Hub:    h,
		RoomID: roomID,
		UserID: userID,
		Send:   make(chan []byte, 256),
	}
	client.Hub.Register <- client

	log.Printf("Cliente '%s' conectado à sala '%s'", client.UserID, client.RoomID)

	// Inicia os processos de leitura e escrita em goroutines separadas.
	go client.WritePump()
	go client.ReadPump()
}

func main() {
	hub := hub.NewHub()
	go hub.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	log.Println("Servidor de Chat iniciado na porta :8084")
	if err := http.ListenAndServe(":8084", nil); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}

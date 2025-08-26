package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/FelipeFelipeRenan/goverse/chat-service/internal/auth"
	"github.com/FelipeFelipeRenan/goverse/chat-service/internal/hub"
	"github.com/FelipeFelipeRenan/goverse/chat-service/pkg/database"
	"github.com/FelipeFelipeRenan/goverse/chat-service/pkg/logger"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// IMPORTANTE: Em produção, valide a origem para permitir apenas seu frontend.
		// Ex: return r.Header.Get("Origin") == "http://localhost:4200"
		return true
	},
}

func serveWs(h *hub.Hub, w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie("access_token")
	if err != nil {
		logger.Error("Erro ao obter cookie de acesso","err", err)
		http.Error(w, "Cookie de autenticação não encontrado", http.StatusUnauthorized)
		return
	}

	tokenString := cookie.Value
	claims, err := auth.ValidateToken(tokenString)
	if err != nil {
		logger.Error("Erro na validação do token","err", err)
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
		logger.Error("Erro no upgrade da conexão para websocket","err", err)
		return
	}

	client := &hub.Client{
		Conn:     conn,
		Hub:      h,
		RoomID:   roomID,
		UserID:   claims.UserID,
		Username: claims.UserName,
		Send:     make(chan []byte, 256),
	}
	client.Hub.Register <- client

	logger.Info("Cliente '%s' (ID: %s) conectado à sala '%s'", client.Username, client.UserID, client.RoomID)

	// Inicia os processos de leitura e escrita em goroutines separadas.
	go client.WritePump()
	go client.ReadPump()
}

func main() {

	logger.Init("INFO", "chat-service")
	dbPool, err := database.Connect()
	if err != nil {
		logger.Error("Erro ao conectar ao banco de dados", "err", err)
	}
	defer dbPool.Close()

	if err := database.RunMigration(dbPool); err != nil {
		logger.Error("Erro ao executar migração", "err", err)

	}
	hub := hub.NewHub()
	go hub.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	port := os.Getenv("CHAT_SERVICE_PORT")
	if port == "" {
		port = "8084"
	}

	server := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	go func() {
		logger.Info("Serviço de Chat rodando", "port", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Erro ao iniciar servidor HTTP", "err", err)
		}
	}()

	// Espera sinal de encerramento (CTRL+C ou kill)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Info("Encerrando Chat Service...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Erro ao encerrar servidor HTTP", "err", err)
	}
}

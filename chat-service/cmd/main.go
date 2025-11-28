package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpHandler "github.com/FelipeFelipeRenan/goverse/chat-service/internal/message/delivery/httpa"
	wsHandler "github.com/FelipeFelipeRenan/goverse/chat-service/internal/message/delivery/websocket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/FelipeFelipeRenan/goverse/chat-service/internal/client"
	"github.com/FelipeFelipeRenan/goverse/chat-service/internal/hub"
	"github.com/FelipeFelipeRenan/goverse/chat-service/internal/message/repository"
	"github.com/FelipeFelipeRenan/goverse/chat-service/internal/message/service"
	"github.com/FelipeFelipeRenan/goverse/chat-service/pkg/redis"
	"github.com/FelipeFelipeRenan/goverse/common/pkg/database"
	"github.com/FelipeFelipeRenan/goverse/common/pkg/logger"
)

func main() {

	logger.Init("INFO", "chat-service")
	dbPool, err := database.Connect()
	if err != nil {
		logger.Error("Erro ao conectar ao banco de dados", "err", err)
	}
	defer dbPool.Close()

	roomGrpcHost := os.Getenv("ROOM_SERVICE_GRPC_HOST")
	if roomGrpcHost == "" {
		roomGrpcHost = "room-service"
	}

	roomGrpcPort := os.Getenv("ROOM_SERVICE_GRPC_PORT")
	if roomGrpcPort == "" {
		roomGrpcPort = ":50052"
	}

	conn, err := grpc.NewClient(roomGrpcHost+roomGrpcPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("falha ao conectar ao room-service via gRPC: ", "err", err)
	}

	defer conn.Close()

	roomClient := client.NewRoomClient(conn)

	redisClient := redis.Init()

	messageRepo := repository.NewMessageRepository(dbPool)
	messageSvc := service.NewMessageService(messageRepo)

	hub := hub.NewHub(messageSvc, redisClient)
	go hub.Run()

	websocketHandler := wsHandler.NewWebSocketHandler(hub, roomClient)
	restHandler := httpHandler.NewMessageHandler(messageSvc, roomClient)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", websocketHandler.ServeWs) // WebSocket
	mux.HandleFunc("GET /rooms/{roomId}/messages", restHandler.GetMessagesByRoom)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	port := os.Getenv("CHAT_SERVICE_PORT")
	if port == "" {
		port = "8084"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
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

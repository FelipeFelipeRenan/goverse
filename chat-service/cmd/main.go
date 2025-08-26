package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/FelipeFelipeRenan/goverse/chat-service/internal/hub"
	"github.com/FelipeFelipeRenan/goverse/chat-service/internal/message/delivery/websocket"
	"github.com/FelipeFelipeRenan/goverse/chat-service/internal/message/repository"
	"github.com/FelipeFelipeRenan/goverse/chat-service/internal/message/service"
	"github.com/FelipeFelipeRenan/goverse/chat-service/pkg/database"
	"github.com/FelipeFelipeRenan/goverse/chat-service/pkg/logger"
)

func main() {

	logger.Init("INFO", "chat-service")
	dbPool, err := database.Connect()
	if err != nil {
		logger.Error("Erro ao conectar ao banco de dados", "err", err)
	}
	defer dbPool.Close()

	// 2. Execução da Migração com Lógica de Retry
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		err = database.RunMigration(dbPool)
		if err == nil {
			break // Sucesso, sai do loop
		}
		logger.Warn("Falha ao executar migração, tentando novamente...", "tentativa", i+1, "erro", err)
		time.Sleep(5 * time.Second) // Espera 5 segundos antes da próxima tentativa
	}
	if err != nil {
		log.Fatalf("Não foi possível executar a migração do banco de dados após %d tentativas: %v", maxRetries, err)
	}

	messageRepo := repository.NewMessageRepository(dbPool)
	messageSvc := service.NewMessageService(messageRepo)

	hub := hub.NewHub(messageSvc)
	go hub.Run()

	wsHandler := websocket.NewWebSocketHandler(hub)

	http.HandleFunc("/ws", wsHandler.ServeWs)

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

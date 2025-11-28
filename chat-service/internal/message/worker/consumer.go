package worker

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/FelipeFelipeRenan/goverse/chat-service/internal/message/domain"
	"github.com/FelipeFelipeRenan/goverse/chat-service/internal/message/service"
	"github.com/FelipeFelipeRenan/goverse/common/pkg/logger"
	"github.com/segmentio/kafka-go"
)

type MessageConsumer struct {
	reader *kafka.Reader
	svc    service.MessageService
}

func NewMessageConsumer(topic string, svc service.MessageService) *MessageConsumer {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{brokers},
		GroupID:  "chat-service-persistence-group", // Importante para load balancing
		Topic:    topic,
		MinBytes: 10e3, // 10 KB
		MaxBytes: 10e6, // 10 MB
	})

	return &MessageConsumer{
		reader: reader,
		svc:    svc,
	}
}

func (c *MessageConsumer) Start(ctx context.Context) {
	logger.Info("Iniciando Kafka Consumer do Chat Service")

	for {
		// le a mensagem (bloqueante)
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			logger.Error("Erro ao ler mensagem do Kafka", "err", err)
			// verifica se o contexto foi cancelado
			if ctx.Err() != nil {
				return
			}
			time.Sleep(1 * time.Second) // Backoff simples
			continue
		}

		// Processa a mensagem
		var msg domain.Message
		if err := json.Unmarshal(m.Value, &msg); err != nil {
			logger.Error("Erro ao fazer unmarshal da mensagem Kafka", "err", err)
			continue
		}

		// Salva no banco (pode ser feito um buffer local para fazer bulk insert)
		// mas por enquanto vai ser salva um a um, porem desacoplado do websocket
		if err := c.svc.ProcessAndSaveMessage(ctx, &msg); err != nil {
			logger.Error("Erro ao salvar mensagem do Kafka no Postgres", "err", err)
			// TODO: Implementar estrategia de retry ou dead lettter queue, para caso ocorrer algum erro com o banco

		} else {
			logger.Debug("Mensagem persistida com successo", "msg_id", msg.ID)
		}

	}
}

func (c *MessageConsumer) Close() error {
	return c.reader.Close()
}

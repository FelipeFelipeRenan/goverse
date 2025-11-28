package kafka

import (
	"context"
	"os"
	"time"

	"github.com/FelipeFelipeRenan/goverse/common/pkg/logger"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(topic string) *Producer {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092" // fallback para dev local sem docker
	}

	writer := &kafka.Writer{
		Addr:                   kafka.TCP(brokers),
		Topic:                  topic,
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
		// Otimização de Batch/lotes
		BatchSize:    100,                   // Envia se juntar 100 mensagens
		BatchTimeout: 50 * time.Microsecond, // ou se passar 50ms
		Async:        true,                  // Não bloqueia a goroutine chamadora
	}

	logger.Info("Kafka Producer inicializado", "brokers", brokers, "topic", topic)
	return &Producer{writer: writer}
}

func (p *Producer) WriteMessage(ctx context.Context, key, value []byte) error {

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: value,
		Time:  time.Now(),
	})
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

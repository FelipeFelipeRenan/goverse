package redis

import (
	"context"
	"os"

	"github.com/FelipeFelipeRenan/goverse/common/pkg/logger"
	"github.com/redis/go-redis/v9"
)

func Init() *redis.Client {

	redisAddr := os.Getenv("REDIS_ADDR")

	if redisAddr == "" {
		redisAddr = "localhost:6379" // padrao para dev local
	}

	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		logger.Info("Não foi possivel conectar ao Redis", "err", err)
		// talvez um panic aqui
	} else {
		logger.Info("Conectado ao Redis com sucesso")
	}

	return client
}

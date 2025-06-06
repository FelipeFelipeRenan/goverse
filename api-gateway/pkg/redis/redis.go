package redis

import (
	"context"
	"os"

	"github.com/FelipeFelipeRenan/goverse/api-gateway/pkg/logger"
	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func Init() {
	Client = redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	ctx := context.Background()

	if err := Client.Ping(ctx).Err(); err != nil {
		logger.Error.Error("Erro no redis", "err", err)
	}
}

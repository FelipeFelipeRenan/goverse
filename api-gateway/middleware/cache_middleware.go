package middleware

import (
	"bytes"
	"context"

	"net/http"
	"time"

	"github.com/FelipeFelipeRenan/goverse/api-gateway/pkg/redis"
)

func CacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.Background()
		key := "cache:" + r.URL.RequestURI()

		cached, err := redis.Client.Get(ctx, key).Result()
		if err == nil {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(cached))
			return
		}

		rw := &responseWriterCache{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(rw, r)

		if rw.status == http.StatusOK {
			_ = redis.Client.Set(ctx, key, rw.body.String(), 300*time.Second).Err()

		}

	})
}

type responseWriterCache struct {
	http.ResponseWriter
	body   bytes.Buffer
	status int
}

func (rw *responseWriterCache) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriterCache) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

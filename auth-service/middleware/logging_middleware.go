package middleware

import (
	"context"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/FelipeFelipeRenan/goverse/auth-service/pkg/logger"
)

type contextKey string

const requestIDKey contextKey = "requestID"

func generateRequestID() string {
	return strconv.FormatInt(rand.Int63(), 16)
}

func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		requestID := generateRequestID()

		ctx := context.WithValue(r.Context(), requestIDKey, requestID)

		r = r.WithContext(ctx)

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r)

		duration := time.Since(start)

		logger.Info.Info("Requisição HTTP",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration", duration.String(),
			"request_id", requestID,
		)

	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}

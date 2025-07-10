package middleware

import (
	"context"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/FelipeFelipeRenan/goverse/auth-service/pkg/logger"
)

func generateRequestID() string {
	return strconv.FormatInt(rand.Int63(), 16)
}

func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(logger.RequestIDKey).(string); ok {
		return id
	}
	return ""
}

func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		requestID := r.Header.Get("X-Request-ID")
		if requestID == "{uuid}" || requestID == "" {
			requestID = generateRequestID()
		}

		ctx := context.WithValue(r.Context(), logger.RequestIDKey, requestID)
		r = r.WithContext(ctx)

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		duration := time.Since(start)

		logger.WithContext(r.Context()).Info("Requisição HTTP",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration", duration.String(),
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

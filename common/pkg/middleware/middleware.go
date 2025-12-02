package middleware

import (
	"context"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/FelipeFelipeRenan/goverse/common/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// --- Chain ---
type Middleware func(http.HandlerFunc) http.HandlerFunc

func Chain(middlewares ...Middleware) Middleware {
	return func(final http.HandlerFunc) http.HandlerFunc {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}

// --- Logging ---
func Logging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" || requestID == "{uuid}" {
			requestID = strconv.FormatInt(rand.Int63(), 16)
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
	}
}

// --- Metrics ---
// Variáveis globais para métricas devem ser registradas apenas uma vez
var (
	httpReqCount *prometheus.CounterVec
	httpReqDur   *prometheus.HistogramVec
)

func InitMetrics(serviceName string) {
	httpReqCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Número total de requisições HTTP",
		},
		[]string{"service", "method", "path", "status"},
	)

	httpReqDur = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duração das requisições HTTP",
		},
		[]string{"service", "method", "path"},
	)
}

func Metrics(serviceName string) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			
			next.ServeHTTP(rec, r)

			duration := time.Since(start).Seconds()

			if httpReqCount != nil {
				httpReqCount.WithLabelValues(
					serviceName, r.Method, r.URL.Path, strconv.Itoa(rec.status),
				).Inc()
			}
			if httpReqDur != nil {
				httpReqDur.WithLabelValues(
					serviceName, r.Method, r.URL.Path,
				).Observe(duration)
			}
		}
	}
}

// Helper para capturar status code
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}
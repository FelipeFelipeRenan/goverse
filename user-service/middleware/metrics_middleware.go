package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/FelipeFelipeRenan/goverse/user-service/pkg/metrics"
)

func MetricsMiddleware(service string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)

			duration := time.Since(start).Seconds()

			metrics.HTTPRequestCount.WithLabelValues(
				service, r.Method, r.URL.Path, strconv.Itoa(rec.status),
			).Inc()

			metrics.HTTPRequestDuration.WithLabelValues(
				service, r.Method, r.URL.Path,
			).Observe(duration)
		}
	}
}


/*
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (rec *statusRecorder) WriteHeader(code int)  {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}*/
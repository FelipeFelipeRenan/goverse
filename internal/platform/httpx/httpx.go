// Package httpx reúne middlewares HTTP compartilhados. A correlação vem do trace do
// OpenTelemetry (otelhttp é o middleware mais externo), então não há request-id próprio.
package httpx

import (
	"bufio"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Wrap aplica, de fora para dentro: tracing -> recover -> access log.
func Wrap(h http.Handler) http.Handler {
	return otelhttp.NewHandler(Recover(AccessLog(h)), "http.server")
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// Unwrap permite que http.ResponseController alcance o writer original (necessário
// para Hijack/Flush, por exemplo no upgrade de WebSocket).
func (w *statusWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

// Hijack é necessário porque o gorilla/websocket faz type assertion para http.Hijacker
// direto no writer recebido (não usa Unwrap).
func (w *statusWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return http.NewResponseController(w.ResponseWriter).Hijack()
}

func (w *statusWriter) Flush() { _ = http.NewResponseController(w.ResponseWriter).Flush() }

func AccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		slog.InfoContext(r.Context(), "http request",
			"method", r.Method, "path", r.URL.Path, "status", sw.status,
			"duration_ms", time.Since(start).Milliseconds())
	})
}

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.ErrorContext(r.Context(), "panic em handler", "panic", rec, "stack", string(debug.Stack()))
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

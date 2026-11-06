package httpx

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Limiter é um token bucket por chave, em memória (um processo; ver o limite do hub do chat).
// perMinute é a taxa de reposição; burst é quantas tentativas seguidas são toleradas.
type Limiter struct {
	mu     sync.Mutex
	m      map[string]*bucket
	rate   float64 // tokens por segundo
	burst  float64
	now    func() time.Time
	maxLen int
}

type bucket struct {
	tokens float64
	last   time.Time
}

func NewLimiter(perMinute, burst int) *Limiter {
	return &Limiter{m: map[string]*bucket{}, rate: float64(perMinute) / 60, burst: float64(burst), now: time.Now, maxLen: 50_000}
}

// Allow consome um token da chave. false = bloqueado.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	b, ok := l.m[key]
	if !ok {
		if len(l.m) >= l.maxLen {
			l.sweep(now)
		}
		b = &bucket{tokens: l.burst, last: now}
		l.m[key] = b
	} else {
		b.tokens = min(l.burst, b.tokens+now.Sub(b.last).Seconds()*l.rate)
		b.last = now
	}
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// sweep remove chaves que já recuperaram o balde cheio (inativas). Exige l.mu.
func (l *Limiter) sweep(now time.Time) {
	for k, b := range l.m {
		if b.tokens+now.Sub(b.last).Seconds()*l.rate >= l.burst {
			delete(l.m, k)
		}
	}
}

// ClientIP devolve o IP do cliente. Atrás de um proxy confiável (trustProxy), usa a ÚLTIMA
// entrada de X-Forwarded-For, a que o nosso proxy acrescentou: as anteriores o cliente controla.
func ClientIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			if ip := strings.TrimSpace(parts[len(parts)-1]); ip != "" {
				return ip
			}
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// TooManyRequests responde 429 no formato padrão de erro.
func TooManyRequests(w http.ResponseWriter) {
	w.Header().Set("Retry-After", strconv.Itoa(30))
	Error(w, http.StatusTooManyRequests, "muitas tentativas, tente novamente em instantes")
}

// RateLimit limita por IP do cliente.
func RateLimit(l *Limiter, trustProxy bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !l.Allow(ClientIP(r, trustProxy)) {
				TooManyRequests(w)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

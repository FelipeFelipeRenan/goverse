package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLimiterBurstAndRefill(t *testing.T) {
	l := NewLimiter(60, 3) // 1 token/s, rajada de 3
	now := time.Unix(1000, 0)
	l.now = func() time.Time { return now }
	for i := 0; i < 3; i++ {
		if !l.Allow("a") {
			t.Fatalf("tentativa %d deveria passar", i)
		}
	}
	if l.Allow("a") {
		t.Fatal("4ª seguida deveria ser bloqueada")
	}
	if !l.Allow("b") {
		t.Fatal("outra chave não é afetada")
	}
	now = now.Add(2 * time.Second)
	if !l.Allow("a") || !l.Allow("a") {
		t.Fatal("2s depois deveriam voltar 2 tokens")
	}
	if l.Allow("a") {
		t.Fatal("e só 2")
	}
}

func TestLimiterSweepsIdleKeys(t *testing.T) {
	l := NewLimiter(60, 2)
	l.maxLen = 2
	now := time.Unix(0, 0)
	l.now = func() time.Time { return now }
	l.Allow("x")
	l.Allow("y")
	now = now.Add(time.Minute) // x e y recuperaram tudo
	l.Allow("z")
	if len(l.m) != 1 {
		t.Fatalf("esperava varredura das chaves inativas, restaram %d", len(l.m))
	}
}

func TestClientIP(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.9:5555"
	r.Header.Set("X-Forwarded-For", "6.6.6.6, 203.0.113.7")
	if got := ClientIP(r, false); got != "10.0.0.9" {
		t.Fatalf("sem proxy confiável deve ignorar XFF, deu %s", got)
	}
	if got := ClientIP(r, true); got != "203.0.113.7" {
		t.Fatalf("com proxy deve usar a última entrada (a do nosso proxy), deu %s", got)
	}
}

func TestRateLimitMiddleware429(t *testing.T) {
	h := RateLimit(NewLimiter(1, 1), false)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	do := func() int {
		r := httptest.NewRequest("POST", "/", nil)
		r.RemoteAddr = "1.2.3.4:1"
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w.Code
	}
	if do() != 200 || do() != http.StatusTooManyRequests {
		t.Fatal("2ª requisição imediata deveria dar 429")
	}
}

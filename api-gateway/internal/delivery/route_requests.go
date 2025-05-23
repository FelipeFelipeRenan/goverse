package delivery

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

var serviceRoutes = map[string]string{
	"/login":                 "http://auth-service:8081",
	"/oauth/google/login":    "http://auth-service:8081",
	"/oauth/google/callback": "http://auth-service:8081",
	"/user":                  "http://user-service:8080",
	"/users":                 "http://user-service:8080",

}

func RouteRequest(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if r.Header.Get("X-Request-ID") == ""{
		r.Header.Set("X-Reader-ID", uuid.New().String())
	}

	for prefix, target := range serviceRoutes {
		if strings.HasPrefix(path, prefix) {
			proxy := newReverseProxy(target)
			proxy.ServeHTTP(w, r)
			return
		}
	}

	http.Error(w, "Rota não encontrada", http.StatusNotFound)
}

func newReverseProxy(target string) *httputil.ReverseProxy {
	targetURL, err := url.Parse(target)
	if err != nil {
		panic("URL do destino inválida: " + target)
	}

	return httputil.NewSingleHostReverseProxy(targetURL)
}

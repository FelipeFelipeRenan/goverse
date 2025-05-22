package delivery

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func RouteRequest(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	method := r.Method

	var target string

	switch {
	case method == http.MethodPost && path == "/login":
		target = "http://auth-service:8081"
	
	case method == http.MethodGet && path == "/oauth/google/login":
		target = "http://auth-service:8081"


	case method == http.MethodPost && path == "/user":
		target = "http://user-service:8080"

	case method == http.MethodGet && strings.HasPrefix(path, "/user/"):
		target = "http://user-service:8080"

	case method == http.MethodGet && path == "/users":
		target = "http://user-service:8080"

	default:
		http.Error(w, "Rota não encontrada", http.StatusNotFound)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   strings.TrimPrefix(target, "http://"),
	})

	proxy.ServeHTTP(w, r)
}

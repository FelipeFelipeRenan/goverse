package delivery

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/FelipeFelipeRenan/goverse/api-gateway/middleware"
)

type Route struct {
	Method string
	Path   string
	Target string
	Prefix bool
	Public bool
}

var routes = []Route{
	{Method: http.MethodGet, Path: "/user/", Target: "http://user-service:8080", Prefix: true, Public: false},
	{Method: http.MethodPost, Path: "/login", Target: "http://auth-service:8081", Public: true},
	{Method: http.MethodGet, Path: "/oauth/google/login", Target: "http://auth-service:8081", Public: true},
	{Method: http.MethodPost, Path: "/user", Target: "http://user-service:8080", Public: true},
	{Method: http.MethodGet, Path: "/users", Target: "http://user-service:8080", Public: true},
}

var serviceRoutes = map[string]string{
	"/login":                 "http://auth-service:8081",
	"/oauth/google/login":    "http://auth-service:8081",
	"/oauth/google/callback": "http://auth-service:8081",
	"/user":                  "http://user-service:8080",
	"/users":                 "http://user-service:8080",
}
func RouteRequest(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	method := r.Method

	var target string
	var isPublic bool

	for _, route := range routes {
		if route.Method != method {
			continue
		}

		if route.Prefix && strings.HasPrefix(path, route.Path) {
			target = route.Target
			isPublic = route.Public
			break
		}

		if !route.Prefix && path == route.Path {
			target = route.Target
			isPublic = route.Public
			break
		}
	}

	if target == "" {
		http.Error(w, "Rota não encontrada", http.StatusNotFound)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   strings.TrimPrefix(target, "http://"),
	})

	handler := http.Handler(proxy)

	// Apenas rotas privadas passam pelo middleware de autenticação
	if !isPublic {
		handler = middleware.AuthMiddleware(handler)
	}

	handler.ServeHTTP(w, r)
}

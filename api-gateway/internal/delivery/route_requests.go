package delivery

import (
	"net/http"
	"strings"

	"github.com/FelipeFelipeRenan/goverse/api-gateway/internal/proxy"
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
	// Users
	{Method: http.MethodGet, Path: "/user/", Target: "http://user-service:8080", Prefix: true, Public: false},
	{Method: http.MethodPost, Path: "/login", Target: "http://auth-service:8081", Public: true},
	{Method: http.MethodGet, Path: "/oauth/google/login", Target: "http://auth-service:8081", Public: true},
	{Method: http.MethodPost, Path: "/user", Target: "http://user-service:8080", Public: true},
	{Method: http.MethodGet, Path: "/users", Target: "http://user-service:8080", Public: true},

	// Membros (ordem importa!)
	{Method: http.MethodPut, Path: "/rooms/", Target: "http://room-service:8082", Public: false, Prefix: true},    // /rooms/{roomID}/members/{memberID}/role
	{Method: http.MethodDelete, Path: "/rooms/", Target: "http://room-service:8082", Public: false, Prefix: true}, // /rooms/{roomID}/members/{memberID}
	{Method: http.MethodGet, Path: "/rooms/", Target: "http://room-service:8082", Public: false, Prefix: true},    // /rooms/{roomID}/members
	{Method: http.MethodPost, Path: "/rooms/", Target: "http://room-service:8082", Public: false, Prefix: true},   // /rooms/{roomID}/members

	// Salas
	{Method: http.MethodPatch, Path: "/rooms/", Target: "http://room-service:8082", Public: false, Prefix: true},
	{Method: http.MethodDelete, Path: "/rooms/", Target: "http://room-service:8082", Public: false, Prefix: true},
	{Method: http.MethodGet, Path: "/rooms/", Target: "http://room-service:8082", Public: true, Prefix: true},
	{Method: http.MethodPost, Path: "/rooms", Target: "http://room-service:8082", Public: false},
	{Method: http.MethodGet, Path: "/rooms", Target: "http://room-service:8082", Public: true},
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

	proxyHandler := proxy.NewReverseProxy(target)

	handler := http.Handler(proxyHandler)

	// Apenas rotas privadas passam pelo middleware de autenticação
	if !isPublic {
		handler = middleware.AuthMiddleware(handler)
	}

	handler.ServeHTTP(w, r)
}

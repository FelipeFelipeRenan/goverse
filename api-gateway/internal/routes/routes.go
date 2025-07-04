package routes

import (
	"net/http"
	"strings"
)

type Route struct {
	Method string
	Path   string
	Target string
	Prefix bool
	Public bool
}

var Routes = []Route{
	// Auth
	{Method: http.MethodPost, Path: "/auth/login", Target: "http://auth-service:8081", Public: true},
	{Method: http.MethodGet, Path: "/oauth/google/login", Target: "http://auth-service:8081", Public: true},

	// Users
	{Method: http.MethodGet, Path: "/user/rooms", Target: "http://room-service:8082", Public: false, Prefix: true},
	{Method: http.MethodPost, Path: "/user", Target: "http://user-service:8080", Public: true},
	{Method: http.MethodGet, Path: "/user", Target: "http://user-service:8080", Public: true, Prefix: true},
	{Method: http.MethodPut, Path: "/user/me", Target: "http://user-service:8080", Public: false, Prefix: true},
	{Method: http.MethodDelete, Path: "/user/me", Target: "http://user-service:8080", Public: false, Prefix: true},

	// Salas
	{Method: http.MethodGet, Path: "/rooms/mine", Target: "http://room-service:8082", Public: false},
	{Method: http.MethodGet, Path: "/rooms", Target: "http://room-service:8082", Public: true},
	{Method: http.MethodPost, Path: "/rooms", Target: "http://room-service:8082", Public: false},

	// Salas e membros (prefixos)
	{Method: http.MethodGet, Path: "/rooms/", Target: "http://room-service:8082", Public: false, Prefix: true},
	{Method: http.MethodPost, Path: "/rooms/", Target: "http://room-service:8082", Public: false, Prefix: true},
	{Method: http.MethodPatch, Path: "/rooms/", Target: "http://room-service:8082", Public: false, Prefix: true},
	{Method: http.MethodPut, Path: "/rooms/", Target: "http://room-service:8082", Public: false, Prefix: true},
	{Method: http.MethodDelete, Path: "/rooms/", Target: "http://room-service:8082", Public: false, Prefix: true},
}

func IsPublicRoute(r *http.Request) bool {
	path := r.URL.Path
	method := r.Method

	// Primeiro: verificação exata (maior prioridade)
	for _, route := range Routes {
		if !route.Prefix && route.Method == method && route.Path == path {
			return route.Public
		}
	}

	// Depois: verificação por prefixo
	for _, route := range Routes {
		if route.Prefix && route.Method == method && strings.HasPrefix(path, route.Path) {
			return route.Public
		}
	}

	return false
}

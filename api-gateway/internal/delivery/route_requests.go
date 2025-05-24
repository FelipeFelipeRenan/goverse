package delivery

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type Route struct {
	Method string
	Path   string
	Target string
	Prefix bool
}

var routes = []Route{
	{Method: http.MethodPost, Path: "/login", Target: "http://auth-service:8081"},
	{Method: http.MethodGet, Path: "/oauth/google/login", Target: "http://auth-service:8081"},
	{Method: http.MethodPost, Path: "/user", Target: "http://user-service:8080"},
	{Method: http.MethodGet, Path: "/users", Target: "http://user-service:8080"},
	{Method: http.MethodGet, Path: "/user/", Target: "http://user-service:8080", Prefix: true},
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

	for _, route := range routes {
		if route.Method == method {
			if route.Prefix && strings.HasPrefix(path, route.Path) {
				target = route.Target
				break
			} else if !route.Prefix && path == route.Path {
				target = route.Target
				break
			}
		}
	}

	if target == "" {
		http.Error(w, "Rota não encontrada", http.StatusNotFound)
	}

	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   strings.TrimPrefix(target, "http://"),
	})

	if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
		r.Header.Set("X-Request-ID", requestID)

	}

	proxy.ServeHTTP(w, r)

}

package delivery

import (
	"net/http"
	"strings"

	"github.com/FelipeFelipeRenan/goverse/api-gateway/internal/proxy"
	"github.com/FelipeFelipeRenan/goverse/api-gateway/internal/routes"
	"github.com/FelipeFelipeRenan/goverse/api-gateway/middleware"
)

func RouteRequest(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	method := r.Method

	var target string

	for _, route := range routes.Routes {
		if route.Method != method {
			continue
		}

		if route.Prefix && strings.HasPrefix(path, route.Path) {
			target = route.Target
			break
		}

		if !route.Prefix && path == route.Path {
			target = route.Target
			break
		}
	}

	if target == "" {
		http.Error(w, "Rota não encontrada", http.StatusNotFound)
		return
	}

	handler := proxy.NewReverseProxy(target)
	handler = middleware.LoggingMiddleware(handler)
	handler.ServeHTTP(w, r)
}

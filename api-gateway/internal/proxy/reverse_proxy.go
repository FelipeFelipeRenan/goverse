package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/FelipeFelipeRenan/goverse/api-gateway/middleware"
)

func NewReverseProxy(target string) http.Handler {
	url, err := url.Parse(target)
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(url)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// 👇 Adiciona o X-User-ID se existir no contexto
		if userID := req.Context().Value(middleware.UserIDKey); userID != nil {
			strUserID := fmt.Sprintf("%v", userID)

			req.Header.Set("X-User-ID", strUserID) // <- isso aqui deve ser string

		}
	}

	return proxy
}

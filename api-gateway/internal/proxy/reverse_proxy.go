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

	// Desabilita keep-alives, conforme já tinha
	proxy.Transport = &http.Transport{
		DisableKeepAlives: true,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Aqui pegamos o userID do contexto original do request e setamos no header do request que será enviado ao backend
		if userID := r.Context().Value(middleware.UserIDKey); userID != nil {
			r.Header.Set("X-User-ID", fmt.Sprintf("%v", userID))
		}
		proxy.ServeHTTP(w, r)
	})
}

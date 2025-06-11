package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
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

		// Apenas repassa o header X-User-ID que já foi setado no middleware
		// não tenta buscar no contexto, pois ele é perdido
		if userID := req.Header.Get("X-User-ID"); userID != "" {
			req.Header.Set("X-User-ID", userID)
		}
	}

	proxy.Transport = &http.Transport{
		DisableKeepAlives: true,
	}

	return proxy
}

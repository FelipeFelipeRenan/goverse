package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)


func NewReverseProxy(target string) http.Handler  {
	url, err := url.Parse(target)
	if err != nil {
		panic(err)
	}

	return httputil.NewSingleHostReverseProxy(url)
}
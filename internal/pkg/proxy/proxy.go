package proxy

import (
	"io"
	"net/http"
)

type ProxyImpl struct {
}

func NewProxyServer() *ProxyImpl {
	return &ProxyImpl{}
}

func (ps *ProxyImpl) ListenAndServe() error {
	server := http.Server{
		Addr: "127.0.0.1:8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ps.proxyHTTP(w, r)
		}),
	}
	return server.ListenAndServe()
}

func (ps *ProxyImpl) proxyHTTP(w http.ResponseWriter, r *http.Request) {
	r.RequestURI = ""
	r.Header.Del("Proxy-Connection")

	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	// Cookies parses and returns the cookies set in the Set-Cookie headers
	resp.Cookies()
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

package easytls

import (
	"net/http"
	"time"
)

// NewServerHTTP will create a new http.Server, with no TLS settings enabled.  This will accept raw HTTP only.
func NewServerHTTP(Addr string, Handlers []SimpleHandler, Middlewares ...MiddlewareHandler) (*http.Server, error) {
	return NewServerHTTPS(nil, Addr, Handlers, Middlewares...)
}

// NewServerHTTPS will create a new TLS-Enabled http.Server.  This will
func NewServerHTTPS(TLS *TLSBundle, Addr string, Handlers []SimpleHandler, Middlewares ...MiddlewareHandler) (*http.Server, error) {
	tls, err := NewTLSConfig(TLS)
	if err != nil {
		return nil, err
	}
	s := &http.Server{
		Addr:              Addr,
		Handler:           NewRouter(Handlers, Middlewares...),
		TLSConfig:         tls,
		ReadTimeout:       time.Minute * 5,
		ReadHeaderTimeout: time.Second * 30,
		WriteTimeout:      time.Minute * 5,
		IdleTimeout:       time.Minute * 5,
	}

	return s, nil
}

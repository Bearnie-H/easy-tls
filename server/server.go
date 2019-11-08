// Package server implements a simple TLS-enabled HTTP server, along with a number of common helper functions to quickly iterating and creating Web services.
//
//
package server

import (
	tlsbundle "easy-tls/tls-bundle"
	"net/http"
	"time"
)

// SimpleServer is a renaming of the Standard http.Server for this package, to allow the ease-of-use extensions provided here.
type SimpleServer http.Server

// NewServerHTTP will create a new SimpleServer, with no TLS settings enabled.  This will accept raw HTTP only.
func NewServerHTTP(Addr string, Handlers []SimpleHandler, Middlewares ...MiddlewareHandler) (*SimpleServer, error) {
	return NewServerHTTPS(nil, Addr, Handlers, Middlewares...)
}

// NewServerHTTPS will create a new TLS-Enabled SimpleServer.  This will
func NewServerHTTPS(TLS *tlsbundle.TLSBundle, Addr string, Handlers []SimpleHandler, Middlewares ...MiddlewareHandler) (*SimpleServer, error) {
	tls, err := tlsbundle.NewTLSConfig(TLS)
	if err != nil {
		return nil, err
	}
	s := &SimpleServer{
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

package server

import (
	"crypto/tls"
	"log"
	"net/http"
	"time"

	easytls "github.com/Bearnie-H/easy-tls"
	"github.com/gorilla/mux"
)

const (
	// DefaultServerAddr represents the default address to serve on,
	// should a value not be provided.
	DefaultServerAddr string = ":8080"
)

// SimpleServer is the extension to the default http.Server this package provides.
type SimpleServer struct {
	server *http.Server
	router *mux.Router
	logger *log.Logger
	tls    *easytls.TLSBundle
	done   chan struct{}
}

// NewServerHTTP will create a new HTTP-only server which will serve on the
// specified IP:Port address. This has NO TLS settings enabled.
// The server returned from this function only has the default http.ServeMux
// as the Router, so should have a dedicated router registered.
//
// The default address of ":8080" will be used if none is provided.
// Only the first string will be treated as the address.
func NewServerHTTP(Addr ...string) (*SimpleServer, error) {
	return NewServerHTTPS(nil, Addr...)
}

// NewServerHTTPS will create a new HTTPS-only server which will serve on
// the specified IP:Port address. The server returned from this function
// only has the default http.ServeMux as the Router, so should have a
// dedicated router registered. The default address of ":8080" will
// be used if none is provided
func NewServerHTTPS(TLS *easytls.TLSBundle, Addr ...string) (*SimpleServer, error) {

	if len(Addr) == 0 {
		Addr = []string{DefaultServerAddr}
	}

	router := NewDefaultRouter()
	logger := easytls.NewDefaultLogger()

	// Create the TLS settings as defined in the TLSBundle
	tls, err := easytls.NewTLSConfig(TLS)
	if err != nil {
		return nil, err
	}

	// Handle a nil TLS bundle safely
	if TLS == nil {
		TLS = &easytls.TLSBundle{
			Enabled: false,
		}
	}

	Server := &SimpleServer{
		server: &http.Server{
			Addr:      Addr[0],
			TLSConfig: tls,
			ErrorLog:  logger,
			Handler:   router,
		},
		router: router,
		logger: logger,
		tls:    TLS,
		done:   make(chan struct{}),
	}

	return Server, nil
}

// CloneTLSConfig will form a proper clone of the underlying tls.Config.
func (S *SimpleServer) CloneTLSConfig() (*tls.Config, error) {
	return easytls.NewTLSConfig(S.tls)
}

// TLSBundle will return a copy of the underlying TLS Bundle
func (S *SimpleServer) TLSBundle() *easytls.TLSBundle {
	return S.tls
}

// SetTimeouts will set the given timeouts of the Server.
// Set 0 to leave uninitialized.
func (S *SimpleServer) SetTimeouts(ReadTimeout, ReadHeaderTimeout, WriteTimeout, IdleTimeout time.Duration) {

	// Timeout to read the full request
	if ReadTimeout != 0 {
		S.server.ReadTimeout = ReadTimeout
	}

	// Timeout to read the header of a request
	if ReadHeaderTimeout != 0 {
		S.server.ReadHeaderTimeout = ReadHeaderTimeout
	}

	// Timeout to finish the full response
	if WriteTimeout != 0 {
		S.server.WriteTimeout = WriteTimeout
	}

	// How long to keep connections alive (if keep-alives are enabled.)
	if IdleTimeout != 0 {
		S.server.IdleTimeout = IdleTimeout
	}
}

// SetLogger will update the logger used by the server from the default to the
// given output.
func (S *SimpleServer) SetLogger(logger *log.Logger) {
	S.logger = logger
	S.server.ErrorLog = logger
}

// Logger will return the internal logger used by the server.
func (S *SimpleServer) Logger() *log.Logger {
	return S.logger
}

// ListenAndServe will start the SimpleServer, serving HTTPS if enabled,
// or HTTP if not. This will properly wait for the shutdown
// to FINISH before returning.
func (S *SimpleServer) ListenAndServe(NonBlocking ...bool) error {

	S.enableAboutHandler()

	S.logger.Printf("Starting server at [ %s ]", S.Addr())

	var ListenAndServe func() error

	if S.tls == nil || !S.tls.Enabled {
		ListenAndServe = func() error {
			defer func() { <-S.done }()
			if err := S.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				return err
			}
			return nil
		}
	} else {
		ListenAndServe = func() error {
			defer func() { <-S.done }()
			if err := S.server.ListenAndServeTLS(S.tls.KeyPair.Certificate, S.tls.KeyPair.Key); err != nil && err != http.ErrServerClosed {
				return err
			}
			return nil
		}
	}

	// If no arguments are given, run in blocking mode.
	if NonBlocking == nil {
		return ListenAndServe()
	}

	// If any args are given, run in non-blocking mode.
	go ListenAndServe()

	return nil
}

// Shutdown will safely shut down the SimpleServer, returning any errors
func (S *SimpleServer) Shutdown() error {
	defer func() { S.done <- struct{}{}; close(S.done) }()
	S.Logger().Printf("Shutting down server at [ %s ]", S.Addr())
	return S.server.Close()
}

// Addr exposes the underlying IP:Port address of the SimpleServer.
func (S *SimpleServer) Addr() string {
	return S.server.Addr
}

// SetKeepAlives will configure the server for whether or not it
// should use Keep-Alives. True implies to use Keep-Alives,
// and false will disable them.
func (S *SimpleServer) SetKeepAlives(SetTo bool) {
	S.server.SetKeepAlivesEnabled(SetTo)
}

// Router will return a pointer to the underlying router used by the server.
func (S *SimpleServer) Router() *mux.Router {
	return S.router
}

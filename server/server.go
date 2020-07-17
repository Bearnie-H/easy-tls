package server

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"sync"
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

	// The actual http.Server implementation
	*http.Server

	// The router to use to match incoming requests to specific handlers
	router *mux.Router

	// The logger to write all messages to
	logger *log.Logger

	// The (optional) TLS resources to use
	tls *easytls.TLSBundle

	// A channel to signal when Shutdown is fully complete and successful
	done chan struct{}

	mu     *sync.Mutex
	active bool
}

// NewServerHTTP will create a new HTTP-only server which will serve on the
// specified IP:Port address. This has NO TLS settings enabled.
// The server returned from this function only has the default http.ServeMux
// as the Router, so should have a dedicated router registered.
//
// The default address of ":8080" will be used if none is provided.
// Only the first string will be treated as the address.
func NewServerHTTP(Addr ...string) *SimpleServer {
	S, _ := NewServerHTTPS(nil, Addr...)
	return S
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

	Server := &SimpleServer{
		Server: &http.Server{
			Addr:      Addr[0],
			TLSConfig: tls,
			ErrorLog:  logger,
			Handler:   router,
		},
		router: router,
		logger: logger,
		tls:    TLS,
		done:   make(chan struct{}),
		mu:     &sync.Mutex{},
		active: false,
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
		S.Server.ReadTimeout = ReadTimeout
	}

	// Timeout to read the header of a request
	if ReadHeaderTimeout != 0 {
		S.Server.ReadHeaderTimeout = ReadHeaderTimeout
	}

	// Timeout to finish the full response
	if WriteTimeout != 0 {
		S.Server.WriteTimeout = WriteTimeout
	}

	// How long to keep connections alive (if keep-alives are enabled.)
	if IdleTimeout != 0 {
		S.Server.IdleTimeout = IdleTimeout
	}
}

// SetLogger will update the logger used by the server from the default to the
// given output.
func (S *SimpleServer) SetLogger(logger *log.Logger) {
	S.logger = logger
	S.Server.ErrorLog = logger
}

// Logger will return the internal logger used by the server.
func (S *SimpleServer) Logger() *log.Logger {
	return S.logger
}

// ListenAndServe will start the SimpleServer, serving HTTPS if enabled,
// or HTTP if not. This will properly wait for the shutdown
// to FINISH before returning.
func (S *SimpleServer) ListenAndServe() error {

	S.enableAboutHandler()

	S.Logger().Printf("Starting server at [ %s ]", S.Addr())

	var ListenAndServe func() error

	if S.tls == nil || !S.tls.Enabled {
		ListenAndServe = func() error {
			defer func() { <-S.done }()
			if err := S.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				return err
			}
			return nil
		}
	} else {
		ListenAndServe = func() error {
			defer func() { <-S.done }()
			if err := S.Server.ListenAndServeTLS(S.tls.KeyPair.Certificate, S.tls.KeyPair.Key); err != nil && err != http.ErrServerClosed {
				return err
			}
			return nil
		}
	}

	S.mu.Lock()
	S.active = true
	S.mu.Unlock()

	return ListenAndServe()
}

// Serve will serve the SimpleServer at the given Listener, rather than allowing it to build
// its own set.
func (S *SimpleServer) Serve(l net.Listener) error {

	S.mu.Lock()
	S.active = true
	S.mu.Unlock()

	S.Server.Addr = l.Addr().String()

	err := S.Server.Serve(l)
	<-S.done
	return err
}

// Shutdown will safely shut down the SimpleServer, returning any errors
func (S *SimpleServer) Shutdown() error {

	S.mu.Lock()
	if !S.active {
		return nil
	}
	S.mu.Unlock()

	defer func() {
		S.Logger().Printf("Finished shutting down server at [ %s ].", S.Addr())
		S.done <- struct{}{}
		close(S.done)
	}()

	S.Logger().Printf("Shutting down server at [ %s ]...", S.Addr())
	return S.Server.Close()
}

// Addr exposes the underlying local address of the SimpleServer.
func (S *SimpleServer) Addr() string {
	return S.Server.Addr
}

// SetKeepAlives will configure the server for whether or not it
// should use Keep-Alives. True implies to use Keep-Alives,
// and false will disable them.
func (S *SimpleServer) SetKeepAlives(SetTo bool) {
	S.Server.SetKeepAlivesEnabled(SetTo)
}

// Router will return a pointer to the underlying router used by the server.
func (S *SimpleServer) Router() *mux.Router {
	return S.router
}

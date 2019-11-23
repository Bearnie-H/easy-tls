package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	easytls "github.com/Bearnie-H/easy-tls"
	"github.com/gorilla/mux"
)

// SimpleServer represents an extension to the standard http.Server
type SimpleServer struct {
	server           *http.Server
	registeredRoutes []string
	tls              *easytls.TLSBundle
	stopped          atomic.Value
}

// NewServerHTTP will create a new http.Server, with no TLS settings enabled.  This will accept raw HTTP only.
func NewServerHTTP(Addr string) (*SimpleServer, error) {
	return NewServerHTTPS(nil, Addr)
}

// NewServerHTTPS will create a new TLS-Enabled http.Server.  This will accept HTTPS, and fully initialize the server based on the TLSBundle provided.
func NewServerHTTPS(TLS *easytls.TLSBundle, Addr string) (*SimpleServer, error) {

	// Create the TLS settings as defined in the TLSBundle
	tls, err := easytls.NewTLSConfig(TLS)
	if err != nil {
		return nil, err
	}

	// Create the Server
	s := &http.Server{
		Addr:      Addr,
		TLSConfig: tls,
	}

	if TLS != nil {
		return &SimpleServer{
			server: s,
			tls:    TLS,
		}, nil
	}
	return &SimpleServer{
		server: s,
		tls: &easytls.TLSBundle{
			Enabled: false,
		},
	}, nil
}

// SetTimeouts will set the given timeouts of the Server to what is passed.  Set 0 to leave uninitialized.
func (S *SimpleServer) SetTimeouts(ReadTimeout, ReadHeaderTimeout, WriteTimeout, IdleTimeout time.Duration) {

	if ReadTimeout != 0 {
		S.server.ReadTimeout = ReadTimeout
	}

	if ReadHeaderTimeout != 0 {
		S.server.ReadHeaderTimeout = ReadHeaderTimeout
	}

	if WriteTimeout != 0 {
		S.server.WriteTimeout = WriteTimeout
	}

	if IdleTimeout != 0 {
		S.server.IdleTimeout = IdleTimeout
	}
}

// ListenAndServe will start the SimpleServer, serving HTTPS if enabled, or HTTP if not
func (S *SimpleServer) ListenAndServe() error {

	S.stopped.Store(false)
	defer S.stopped.Store(true)

	if S.tls.Enabled {
		log.Printf("Serving HTTPS at: %s\n", S.server.Addr)
		if err := S.server.ListenAndServeTLS(S.tls.KeyPair.Certificate, S.tls.KeyPair.Key); err != nil && err != http.ErrServerClosed {
			return err
		}
	} else {
		log.Printf("Serving HTTP at: %s\n", S.server.Addr)
		if err := S.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}
	}

	for !S.stopped.Load().(bool) {
		log.Println("Waiting for server to shut down...")
		time.Sleep(time.Second)
	}

	return nil
}

// Shutdown will safely shut down the SimpleServer, returning any errors
func (S *SimpleServer) Shutdown() error {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)

	defer cancel()
	defer func() { S.stopped.Store(true) }()

	return S.server.Shutdown(ctx)
}

// RegisterRouter will register the given Handler (typically an *http.ServeMux or *mux.Router) as the http Handler for the server.
func (S *SimpleServer) RegisterRouter(r http.Handler) {
	S.server.Handler = r
}

// EnableAboutHandler will enable and set up the "about" handler, to display the available routes.  This must be the last route registered in order for the full set of routes to be displayed.
func (S *SimpleServer) EnableAboutHandler(r *mux.Router) {
	routeList := append([]string{fmt.Sprintf("%-50s|    %v", "/about", []string{http.MethodGet})}, S.registeredRoutes...)
	routes := strings.Join(routeList, "\n")
	aboutHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(routes + "\n"))
	}
	r.HandleFunc("/about", aboutHandler)
}

// Addr exposes the underlying TCP address of the SimpleServer.
func (S *SimpleServer) Addr() string {
	return S.server.Addr
}

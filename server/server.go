package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	easytls "github.com/Bearnie-H/easy-tls"
	"github.com/gorilla/mux"
)

// SimpleServer is the extension to the default http.Server this package provides.
type SimpleServer struct {
	server              *http.Server
	registeredRoutes    []string
	tls                 *easytls.TLSBundle
	stopped             atomic.Value
	aboutHandlerEnabled bool
}

// NewServerHTTP will create a new HTTP-only server which will serve on the specified IP:Port address.  This has NO TLS settings enabled.  The server returned from this function only has the default http.ServeMux as the Router, so should have a dedicated router registered.
//
// The default address of ":8080" will be used if none is provided
func NewServerHTTP(Addr string) (*SimpleServer, error) {

	if Addr == "" {
		Addr = ":8080"
	}

	return NewServerHTTPS(nil, Addr)
}

// NewServerHTTPS will create a new HTTPS-only server which will serve on the specified IP:Port address.  The server returned from this function only has the default http.ServeMux as the Router, so should have a dedicated router registered.
//
// The default address of ":8080" will be used if none is provided
func NewServerHTTPS(TLS *easytls.TLSBundle, Addr string) (*SimpleServer, error) {

	if Addr == "" {
		Addr = ":8080"
	}

	// If the TLSBundle is nil, just create a server without TLS settings.
	if TLS == nil {
		return &SimpleServer{
			server: &http.Server{
				Addr:      Addr,
				TLSConfig: &tls.Config{},
			},
			tls: &easytls.TLSBundle{
				Enabled: false,
			},
		}, nil
	}

	// Create the TLS settings as defined in the TLSBundle
	tls, err := easytls.NewTLSConfig(TLS)
	if err != nil {
		return nil, err
	}

	// Create the TLS-Enabled server.
	return &SimpleServer{
		server: &http.Server{
			Addr:      Addr,
			TLSConfig: tls,
		},
		tls: TLS,
	}, nil
}

// SetTimeouts will set the given timeouts of the Server.  Set 0 to leave uninitialized.
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

// ListenAndServe will start the SimpleServer, serving HTTPS if enabled, or HTTP if not.  This will properly wait for the shutdown to FINISH before returning.
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

	// Block while waiting for the server to fully shutdown.
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
	S.aboutHandlerEnabled = true
	routeList := append([]string{fmt.Sprintf("%-50s|    %v", "/about", []string{http.MethodGet})}, S.registeredRoutes...)
	routes := strings.Join(routeList, "\n")
	aboutHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(routes + "\n"))
	}
	r.HandleFunc("/about", aboutHandler)

	r.NotFoundHandler = http.RedirectHandler("/about", http.StatusMovedPermanently)
	r.MethodNotAllowedHandler = http.RedirectHandler("/about", http.StatusMovedPermanently)
}

// Addr exposes the underlying IP:Port address of the SimpleServer.
func (S *SimpleServer) Addr() string {
	return S.server.Addr
}

// SetKeepAlives will configure the server for whether or not it should use Keep-Alives.  True implies to use Keep-Alives, and false will disable them.
func (S *SimpleServer) SetKeepAlives(SetTo bool) {
	S.server.SetKeepAlivesEnabled(SetTo)
}

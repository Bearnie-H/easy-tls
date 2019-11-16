package easytls

import (
	"context"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
)

// SimpleServer represents an extension to the standard http.Server
type SimpleServer struct {
	server           *http.Server
	registeredRoutes []string
	tls              *TLSBundle
	stopped          atomic.Value
}

// NewServerHTTP will create a new http.Server, with no TLS settings enabled.  This will accept raw HTTP only.
func NewServerHTTP(Addr string) (*SimpleServer, error) {
	return NewServerHTTPS(nil, Addr)
}

// NewServerHTTPS will create a new TLS-Enabled http.Server.  This will accept HTTPS, and fully initialize the server based on the TLSBundle provided.
func NewServerHTTPS(TLS *TLSBundle, Addr string) (*SimpleServer, error) {

	// Create the TLS settings as defined in the TLSBundle
	tls, err := NewTLSConfig(TLS)
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
		tls: &TLSBundle{
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
	var err error
	if S.tls.Enabled {
		log.Printf("Serving HTTPS at: %s\n", S.server.Addr)
		err = S.server.ListenAndServeTLS(S.tls.KeyPairs[0].Certificate, S.tls.KeyPairs[0].Key)
	} else {
		log.Printf("Serving HTTP at: %s\n", S.server.Addr)
		err = S.server.ListenAndServe()
	}
	for !S.stopped.Load().(bool) {
		log.Println("Waiting for server to shut down...")
		time.Sleep(time.Second)
	}
	if err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown will safely shut down the SimpleServer, returning any errors
func (S *SimpleServer) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	defer func() { S.stopped.Store(true) }()
	S.server.Shutdown(ctx)
}

// RegisterRouter will register the given Handler (typically an *http.ServeMux or *mux.Router) as the http Handler for the server.
func (S *SimpleServer) RegisterRouter(r http.Handler) {
	S.server.Handler = r
}

// EnableAboutHandler will enable and set up the "about" handler, to display the available routes.  This must be the last route registered in order for the full set of routes to be displayed.
func (S *SimpleServer) EnableAboutHandler(r *mux.Router) {
	routeList := strings.Join(S.registeredRoutes, "\n")
	aboutHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(routeList))
	}
	r.HandleFunc("/about", aboutHandler)
}

// Addr exposes the underlying TCP address of the SimpleServer.
func (S *SimpleServer) Addr() string {
	return S.server.Addr
}

// ConfigureReverseProxy will convert a freshly created SimpleServer into a ReverseProxy, forwarding all incoming traffic based on the RouteMatcher func provided.  This will create the necessary HTTP handler, and configure the necessary routing.
//
// PathPrefix is variadic to allow for no argument to be specified.  If no argument is specified, this will forward all traffic starting with path "/".  If multiple PathPrefix arguments are provided, only the first will be used.
func (S *SimpleServer) ConfigureReverseProxy(Client *SimpleClient, RouteMatcher ReverseProxyRouterFunc, PathPrefix ...string) {
	r := NewDefaultRouter()

	p := "/"
	if len(PathPrefix) > 1 {
		p = PathPrefix[0]

	}
	r.PathPrefix(p).HandlerFunc(doReverseProxy(Client, Client.IsTLS(), RouteMatcher))
	AddMiddlewares(r, MiddlewareDefaultLogger)
	S.RegisterRouter(r)
}

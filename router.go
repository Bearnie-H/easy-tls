package easytls

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// NewDefaultRouter will initialize a new HTTP Router, based on the gorilla/mux implementation.
func NewDefaultRouter() *mux.Router {
	r := mux.NewRouter()

	// Don't be pedantic about possible trailing slashes in the routes.
	r.StrictSlash(true)

	// Register some default handlers
	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	r.MethodNotAllowedHandler = http.HandlerFunc(methodNotAllowedHandler)

	return r
}

// NewRouter will build a new complex router, with the given routes, and middlewares.  More can be added later, if needed.
func NewRouter(s *SimpleServer, Handlers []SimpleHandler, Middlewares ...MiddlewareHandler) *mux.Router {
	r := NewDefaultRouter()

	AddMiddlewares(r, Middlewares...)

	AddHandlers(false, s, r, Handlers...)

	return r
}

// AddMiddlewares is a convenience wrapper for the mux.Router "Use" function
func AddMiddlewares(r *mux.Router, middlewares ...MiddlewareHandler) {
	for _, mwf := range middlewares {
		r.Use(mux.MiddlewareFunc(mwf))
	}
}

// AddHandlers will add the given handlers to the router, with the verbose flag determining if a log message should be generated for each added route.
func AddHandlers(verbose bool, s *SimpleServer, r *mux.Router, Handlers ...SimpleHandler) {
	// Register the routes, this IS order dependent.
	for _, Node := range Handlers {
		r.Handle(Node.Path, Node.Handler).Methods(Node.Methods...)
		if verbose {
			log.Printf("Registered route %s with accepted methods %v", Node.Path, Node.Methods)
		}
		s.registeredRoutes = append(s.registeredRoutes, fmt.Sprintf("%-30s|    %v", Node.Path, Node.Methods))
	}
}

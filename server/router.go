package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// NewDefaultRouter will create a new Router, based on the gorilla/mux implementation.  This will pre-set the "trailing-slash" behaviour to not be pedantic, as well as initialing the "Not-Found" and "Method-Not-Allowed" behaviours to simply return the corresponding status codes.  These can be overridden if desired.
func NewDefaultRouter() *mux.Router {
	r := mux.NewRouter()

	// Don't be pedantic about possible trailing slashes in the routes.
	r.StrictSlash(true)

	// Register some default handlers
	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	r.MethodNotAllowedHandler = http.HandlerFunc(methodNotAllowedHandler)

	return r
}

// NewRouter will create a new Router with ALL of the specified Routes and Middlewares, and update the SimpleServer with the registration information.  This will NOT register the router, as more handlers and/or middlewares can be added following this function if necessary.
func NewRouter(s *SimpleServer, Handlers []SimpleHandler, Middlewares ...MiddlewareHandler) *mux.Router {

	if s == nil {
		panic(errors.New("easytls router error - SimpleServer not initialized"))
	}

	r := NewDefaultRouter()

	AddMiddlewares(r, Middlewares...)

	AddHandlers(false, s, r, Handlers...)

	return r
}

// AddMiddlewares is a convenience wrapper for the mux.Router "Use" function.  This will add the middlewares to the router in the order specified (which also defines their execution order).
func AddMiddlewares(r *mux.Router, middlewares ...MiddlewareHandler) {
	for _, mwf := range middlewares {
		r.Use(mux.MiddlewareFunc(mwf))
	}
}

// AddHandlers will add the given handlers to the router, with the verbose flag determining if a log message should be generated for each added route.
func AddHandlers(verbose bool, s *SimpleServer, r *mux.Router, Handlers ...SimpleHandler) {
	if s.aboutHandlerEnabled {
		panic(errors.New("easytls server error - Registered new routes after \"/about\" route"))
	}
	// Register the routes, this IS order dependent.
	for _, Node := range Handlers {
		r.Handle(Node.Path, Node.Handler).Methods(Node.Methods...)
		if verbose {
			log.Printf("Registered route %s with accepted methods %v", Node.Path, Node.Methods)
		}
		s.registeredRoutes = append(s.registeredRoutes, fmt.Sprintf("%-50s|    %v", Node.Path, Node.Methods))
	}
}

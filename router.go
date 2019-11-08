package easytls

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// NewRouter will initialize a new HTTP Router, based on the gorilla/mux implementation.
func NewRouter(Handlers []SimpleHandler, Middlewares ...MiddlewareHandler) *mux.Router {
	r := mux.NewRouter()

	// Don't be pedantic about possible trailing slashes in the routes.
	r.StrictSlash(true)

	// Register some default handlers
	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	r.MethodNotAllowedHandler = http.HandlerFunc(methodNotAllowedHandler)

	// Register the middlwares, this IS order dependent.
	r.Use(defaultLoggingMiddleware)
	for _, middleware := range Middlewares {
		r.Use(mux.MiddlewareFunc(middleware))
	}

	// Register the routes, this IS order dependent.
	for _, Node := range Handlers {
		r.Handle(Node.Path, Node.Handler).Methods(Node.Methods...)
		log.Printf("Registered route %s with accepted methods %v", Node.Path, Node.Methods)
	}

	return r
}

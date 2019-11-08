package server

import "net/http"

// SimpleHandler represents a simplification to the standard http handlerFuncs, allowing simpler registration and logging with Routers.
type SimpleHandler struct {
	Handler http.Handler
	Path    string
	Methods []string
}

// MiddlewareHandler represents the general format for a Middleware handler.
type MiddlewareHandler func(http.Handler) http.Handler

// notFoundHandler represents the function to call if the path requested has not been registered.
func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

// methodNotAllowedHandler represents the function to call if the method used on the path has not been registered explicitly.
func methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

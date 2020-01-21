package server

import (
	"net/http"
	"strings"
)

// SimpleHandler represents a simplification to the standard http handlerFuncs, allowing simpler registration and logging with Routers.
type SimpleHandler struct {
	Handler http.HandlerFunc
	Path    string
	Methods []string
}

// NewSimpleHandler will create and return a new SimpleHandler, ready to be used.
func NewSimpleHandler(h http.HandlerFunc, Path string, Methods ...string) SimpleHandler {

	if !strings.HasPrefix(Path, "/") {
		Path = "/" + Path
	}

	return SimpleHandler{
		Handler: h,
		Path:    Path,
		Methods: Methods,
	}
}

// notFoundHandler represents the default function to call if the path requested has not been registered.
func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

// methodNotAllowedHandler represents the default function to call if the method used on the path has not been registered explicitly.
func methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

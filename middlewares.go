package easytls

import (
	"log"
	"net/http"
)

// MiddlewareHandler represents the general format for a Middleware handler.
type MiddlewareHandler func(http.Handler) http.Handler

// DefaultLoggingMiddleware provides a simple logging middleware, to view incoming connections as they arrive.
func DefaultLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Recieved [%s] Request for URL \"%s\" from Address: [%s].\n", r.Method, r.URL.String(), r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

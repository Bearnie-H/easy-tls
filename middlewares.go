package easytls

import (
	"log"
	"net/http"
	"time"
)

// MiddlewareHandler represents the general format for a Middleware handler.
type MiddlewareHandler func(http.Handler) http.Handler

// MiddlewareDefaultLogger provides a simple logging middleware, to view incoming connections as they arrive.
func MiddlewareDefaultLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Recieved [ %s ] [ %s ] Request for URL \"%s\" from Address: [ %s ].\n", r.Proto, r.Method, r.URL.String(), r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

// MiddlewareLimitMaxConnections will provide a mechanism to strictly limit the maximum number of connections served.
func MiddlewareLimitMaxConnections(ConnectionLimit int) func(http.Handler) http.Handler {
	semaphore := make(chan struct{}, ConnectionLimit)

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			h.ServeHTTP(w, r)
		})
	}

}

// MiddlewareLimitConnectionRate will limit the server to a maximum of one connection per CycleTime.
func MiddlewareLimitConnectionRate(CycleTime time.Duration) func(http.Handler) http.Handler {
	ticker := time.NewTicker(CycleTime)

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			<-ticker.C
			h.ServeHTTP(w, r)
		})
	}
}

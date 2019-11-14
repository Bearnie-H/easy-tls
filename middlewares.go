package easytls

import (
	"log"
	"net/http"
	"time"
)

// MiddlewareHandler represents the general format for a Middleware handler.
type MiddlewareHandler func(http.Handler) http.Handler

// DefaultLoggingMiddleware provides a simple logging middleware, to view incoming connections as they arrive.
func DefaultLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Recieved [ %s ] [ %s ] Request for URL \"%s\" from Address: [ %s ].\n", r.Proto, r.Method, r.URL.String(), r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

// LimitMaxConnectionMiddleware will provide a mechanism to strictly limit the maximum number of connections served.
func LimitMaxConnectionMiddleware(next http.Handler, ConnectionLimit int) http.Handler {
	semaphore := make(chan struct{}, ConnectionLimit)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		semaphore <- struct{}{}
		defer func() { <-semaphore }()
		next.ServeHTTP(w, r)
	})
}

// LimitMaxConnectionRate will limit the server to a maximum of one connection per CycleTime.
func LimitMaxConnectionRate(next http.Handler, CycleTime time.Duration) http.Handler {
	ticker := time.NewTicker(CycleTime)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-ticker.C
		next.ServeHTTP(w, r)
	})
}

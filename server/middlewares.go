package server

import (
	"log"
	"net/http"
	"time"
)

// MiddlewareHandler represents the Type which must be satified by any
// function to be used as a middleware function in the Server chain.
type MiddlewareHandler = func(http.Handler) http.Handler

// MiddlewareDefaultLogger provides a simple logging middleware, to view
// incoming connections as they arrive and print a basic set of
// properties of the request.
func MiddlewareDefaultLogger(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Printf("[MiddlewareDefaultLogger] Recieved [ %s ] [ %s ] Request for URL \"%s\" from Address: [ %s ].\n", r.Proto, r.Method, r.URL.String(), r.RemoteAddr)
			next.ServeHTTP(w, r)
		})
	}
}

// MiddlewareLimitMaxConnections will provide a mechanism to strictly limit the
// maximum number of concurrent requests served. Verbose mode includes a log
// message when a request begins processing through this function. If the
// request is not processed within Timeout, a failed statusCode will
// be generated and sent back.
func MiddlewareLimitMaxConnections(ConnectionLimit int, Timeout time.Duration, logger *log.Logger) func(http.Handler) http.Handler {
	semaphore := make(chan struct{}, ConnectionLimit)

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Set up a timer that will "tick" at some point in the future
			if Timeout == 0 {
				Timeout = time.Hour * 6
			}
			timer := time.NewTimer(Timeout)

			// Block on one of two channels...
			select {

			// If there is room for a request, stop the timer and process it.
			case semaphore <- struct{}{}:
				if !timer.Stop() {
					<-timer.C
				}
				defer func() { <-semaphore }()
				if logger != nil {
					logger.Printf("[MiddlewareLimitMaxConnections] Processing [ %s ] [ %s ] Request for URL \"%s\" from Address: [ %s ].\n", r.Proto, r.Method, r.URL.String(), r.RemoteAddr)
				}
				h.ServeHTTP(w, r)

				// If the timer expires, write a timeout response and exit
			case <-timer.C:
				timer.Stop()
				w.WriteHeader(http.StatusRequestTimeout)
				if logger != nil {
					logger.Printf("[MiddlewareLimitMaxConnections] [ %s ] [ %s ] Request for URL \"%s\" from Address: [ %s ] - Timeout\n", r.Proto, r.Method, r.URL.String(), r.RemoteAddr)
				}
			}
		})
	}

}

// MiddlewareLimitConnectionRate will limit the rate at which the Server will
// process incoming requests. This will process no more than 1 request per
// CycleTime. Verbose mode includes a log message when a request begins
// processing through this function. If the request is not processed within
// Timeout, a failed statusCode will be generated and sent back.
func MiddlewareLimitConnectionRate(CycleTime time.Duration, Timeout time.Duration, logger *log.Logger) func(http.Handler) http.Handler {
	ticker := time.NewTicker(CycleTime)

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Set up a timer that will "tick" at some point in the future
			if Timeout == 0 {
				Timeout = time.Hour * 6
			}
			timer := time.NewTimer(Timeout)

			// Block on one of two channels...
			select {

			// If there is room for a request, stop the timer and process it.
			case <-ticker.C:
				if !timer.Stop() {
					<-timer.C
				}
				if logger != nil {
					log.Printf("[MiddlewareLimitConnectionRate] Processing [ %s ] [ %s ] Request for URL \"%s\" from Address: [ %s ].\n", r.Proto, r.Method, r.URL.String(), r.RemoteAddr)
				}
				h.ServeHTTP(w, r)

				// If the timer expires, write a timeout response and exit
			case <-timer.C:
				w.WriteHeader(http.StatusRequestTimeout)
				if logger != nil {
					log.Printf("[MiddlewareLimitConnectionRate] [ %s ] [ %s ] Request for URL \"%s\" from Address: [ %s ] - Timeout\n", r.Proto, r.Method, r.URL.String(), r.RemoteAddr)
				}
				timer.Stop()
			}
		})
	}
}

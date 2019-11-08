package server

import (
	"log"
	"net/http"
)

func defaultLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Recieved [%s] Request for URL \"%s\" from Address: [%s].\n", r.Method, r.URL.String(), r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

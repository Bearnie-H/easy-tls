package main

import (
	"fmt"
	"net/http"
)

// Module-specific HTTP handlers go here...

// Template function definition to use. Copy and update the name to simplify creating new
// Handlers.
func template() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// ...
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(fmt.Sprintf("Handler for [ %s ] request to [ %s ] is not yet implemented", r.Method, r.URL.Path)))
	})
}

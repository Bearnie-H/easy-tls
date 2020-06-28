package main

import (
	"net/http"
)

// Module-specific HTTP handlers go here...

// Template function definition to use.
func template() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ...
	})
}

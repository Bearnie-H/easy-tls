package main

import "net/http"

// Plugins must supply their own middlewares if they want to have additional
// specific logic applied to the Handlers it returns. These middleware(s)
// will only be applied to the handlers and routes presented by the plugin
// itself, if you wish to have middlewares applies to the entire tree, or
// shared by multiple modules, this logic will have to be injected somewhere
// else.
//
// As long as these functions return an http.Handler, they can have any input

// HideOnStop is a simple piece of middleware to allow for a handler to no longer be accessible
// if the overall module has been Stopped()
func HideOnStop(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// If the plugin has been stopped, return a 404
		if Killed.Load().(bool) {
			http.NotFound(w, r)
			return
		}
		// Otherwise, just handle the request
		next.ServeHTTP(w, r)
	})
}

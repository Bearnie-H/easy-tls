package main

import (
	"fmt"
	"net/http"

	"github.com/Bearnie-H/easy-tls/server"
)

// Module-specific HTTP handlers go here...

// Template function definition to use. Copy and update the name to simplify creating new
// Handlers.
func template() server.SimpleHandler {
	return server.SimpleHandler{

		// Match on a specific URL path
		Path: PluginName,

		// Match this on GET requests
		Methods: []string{
			http.MethodGet,
			// ...
		},

		// Require the URL Query parameter "foo" be set to "big", and allow for the optional value "bar" to be any integer
		Queries: []server.QueryKeyValue{
			{Key: "foo", Value: "big"},
			{Key: "bar", Value: "{bar:[0-9]+}"},
			// ...
		},

		// Add a description to this route, to be registered and shown on the "/about" handler if enabled.
		// This functions as an externally visible documentation tool for the server.
		Description: "This route is a template or example route to help illustrate standard and idiomatic handler provisioning with the EasyTLS library.",

		// Finally, supply the handler, either as a pre-defined function or as an anonymous function.
		// Wrap the anonymous function in a piece of middleware to allow this handler to be hidden when
		// the module has Stop() called.
		Handler: HideOnStop(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ThreadCount.Add(1)
			defer ThreadCount.Done()

			// Add your actual handler logic below...
			w.WriteHeader(http.StatusNotImplemented)
			w.Write([]byte(fmt.Sprintf("Handler for [ %s ] request to [ %s ] is not yet implemented", r.Method, r.URL.Path)))
		})),
	}
}

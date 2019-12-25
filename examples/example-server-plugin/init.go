package main

import (
	"fmt"
	"net/http"

	"github.com/Bearnie-H/easy-tls/server"
)

// Init is the function to start the plugin logic.
func Init(args ...interface{}) ([]server.SimpleHandler, error) {

	// Perform the non-specific module initialization steps.
	if err := defaultInitialization(args...); err != nil {
		return nil, fmt.Errorf("easytls module error: Failed to perform standard initialization - %s", err)
	}

	// Perform the module-specific initialization steps.
	if err := moduleInitialization(args...); err != nil {
		return nil, fmt.Errorf("easytls module error: Failed to perform module-specific initialization - %s", err)
	}

	// Add in the handlers for the API nodes, methods, and functions to be handled by this plugin.
	// The routes need to be added longest-first in order to be successfully registered without collisions or name-mixups
	// Furthermore, routes with explicit path components need to be added before routes with variable path components
	//
	//	For Example:
	//
	//		/path/route
	//		/path/{User}
	//
	// Is the correct ordering if there is a route where "route" is explicit and constant, and there is another route
	//  that handles all other values of that path segment.
	h := []server.SimpleHandler{}
	h = append(h, server.SimpleHandler{Handler: GetPluginVersion, Path: fmt.Sprintf("/%s/version", PluginName), Methods: []string{http.MethodGet, http.MethodPost}})
	// ... Append your handlers here

	return h, nil
}

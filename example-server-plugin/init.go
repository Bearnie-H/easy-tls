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
	h := []server.SimpleHandler{}
	h = append(h, server.SimpleHandler{Handler: GetPluginVersion, Path: fmt.Sprintf("/%s/version", PluginName), Methods: []string{http.MethodGet, http.MethodPost}})
	// ... Other necessary handlers here

	return h, nil
}

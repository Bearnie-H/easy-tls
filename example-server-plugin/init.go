package main

import (
	"fmt"
	"net/http"

	easytls "github.com/Bearnie-H/easy-tls"
)

// Init is the function to start the plugin logic.
func Init() ([]easytls.SimpleHandler, error) {

	// Perform the non-specific module initialization steps.
	if err := defaultInitialization(); err != nil {
		return nil, fmt.Errorf("easytls module error: Failed to perform standard initialization - %s", err)
	}

	// Perform the module-specific initialization steps.
	if err := moduleInitialization(); err != nil {
		return nil, fmt.Errorf("easytls module error: Failed to perform module-specific initialization - %s", err)
	}

	// Add in the handlers for the API nodes, methods, and functions to be handled by this plugin.
	// The routes need to be added longest-first in order to be successfully registered without collisions or name-mixups
	h := []easytls.SimpleHandler{}
	h = append(h, easytls.SimpleHandler{Handler: GetPluginVersion, Path: fmt.Sprintf("/%s/version", PluginName), Methods: []string{http.MethodGet, http.MethodPost}})
	// ... Other necessary handlers here

	return h, nil
}

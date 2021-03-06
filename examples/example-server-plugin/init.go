package main

import (
	"fmt"

	"github.com/Bearnie-H/easy-tls/server"
)

// Init is the function to start the plugin logic.
//
// Init can have one of two function signatures and be successfully loaded
// into an EasyTLS framework.
//
//	func(...interface{}) ([]server.SimpleHandler, error)
//
//	func(...interface{}) ([]server.SimpleHandler, string, error)
//
// In all cases, this function MUST return the set of routes exported to the
// framework, and any errors which occured during initialization.
//
// The optional string return value differentiates how the framework will
// process and register the routes internally. If the signature provides this
// string, it MUST represent a common URL prefix for ALL exported routes.
// This must be true, as the framework will create a dedicated sub-router
// matching this exact URL prefix when searching for the handler to accept
// an incoming request. If possible, this should be used, as it provides a
// reasonable efficiency increase in handler dispatching.
//
// Best practices for this prefix value is to simply use the "PluginName"
// global variable, as this is both simple and a meaningfully clear value.
func Init(args ...interface{}) ([]server.SimpleHandler, string, error) {
	ThreadCount.Add(1)
	defer ThreadCount.Done()

	// Perform the non-specific module initialization steps.
	if err := defaultInitialization(args...); err != nil {
		return nil, PluginName, fmt.Errorf("easytls module error: Failed to perform standard initialization - %s", err)
	}

	// Perform the module-specific initialization steps.
	if err := moduleInitialization(args...); err != nil {
		return nil, PluginName, fmt.Errorf("easytls module error: Failed to perform module-specific initialization - %s", err)
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

	// Perform this only once, on the first instance of the module being started.
	Registered.Do(func() {
		// Canonically, the handlers to be returned should be formatted and constructed in the manner of GetPluginVersion.
		// This allows for a full decorator-like pattern for the full set of routing / matching properties, as well
		// as the ability to have the internal handler logic function as a closure over extracted properties.
		// Furthermore, this style also provides room and opportunity for route-specific "middle-ware" which can be selectively
		// applied to only some routes, rather than applying a given middleware to all routes of a module.
		h = append(h, GetPluginVersion())
		// ... Append your handlers here
	})

	return h, PluginName, nil
}

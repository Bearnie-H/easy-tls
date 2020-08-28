package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Bearnie-H/easy-tls/server"
)

// ExitHandler is the generic function to simplify failing out of a HTTP Handler within a plugin
//
// The general use of this is:
//
//	if err := foo(); err != nil {
//		ExitHandler(w, http.StatusInternalServerError, "Failed to foo the bar for ID [ %s ] with index [ %d ]", err, ID, index)
//		return
//	}
//
// This will write the status code to the response, as well as the result of
// fmt.Sprintf(Message, args...) to the response, and WriteStatus(Message, err, false, args...)
// to the plugin status writer.
func ExitHandler(w http.ResponseWriter, StatusCode int, Message string, err error, args ...interface{}) {
	w.WriteHeader(StatusCode)
	w.Write([]byte(fmt.Sprintf(Message, args...)))
	StatusChannel.Printf(Message, err, args...)
	return
}

// GetPluginVersion allows for the version of the Server Plugin to be requested via the RESTful interface
func GetPluginVersion() server.SimpleHandler {
	return server.SimpleHandler{

		// Match this handler for a specific URL Path
		Path: fmt.Sprintf("%s/version", PluginName),

		// Accept GET, HEAD, and POST requests
		Methods: []string{
			http.MethodGet,
			http.MethodHead,
			http.MethodPost,
		},

		// Document what this handler does, and why it's valuable
		Description: "Return the JSON encoded semantic version of this module, to allow other services or utilities to determine their compatibility.",

		// Define the actual logic.
		// The Version-Check handler is purposefully not hidden when the module is
		// Stopped, to ensure that as long as it is loaded, other services can still
		// check and verify whether or not they are compatible if and when the module
		// is started again.
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ThreadCount.Add(1)
			defer ThreadCount.Done()

			if err := json.NewEncoder(w).Encode(PluginVersion); err != nil {
				ExitHandler(w, http.StatusInternalServerError, "Failed to JSON encode version information for module [ %s ]", err, PluginName)
				return
			}
		}),
	}
}

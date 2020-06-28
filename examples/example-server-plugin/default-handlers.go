package main

import (
	"encoding/json"
	"fmt"
	"net/http"
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
func GetPluginVersion() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(PluginVersion); err != nil {
			ExitHandler(w, http.StatusInternalServerError, "Failed to JSON encode version information", err)
			return
		}
	})
}

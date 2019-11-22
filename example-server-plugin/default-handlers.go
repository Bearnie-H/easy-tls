package main

import (
	"encoding/json"
	"net/http"
)

// GetPluginVersion allows for the version of the Server Plugin to be requested via the RESTful interface
func GetPluginVersion(w http.ResponseWriter, r *http.Request) {
	if err := json.NewEncoder(w).Encode(PluginVersion); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

package main

import (
	"github.com/Bearnie-H/easy-tls/client"
	"time"
)

// Main is the top-level ACTION performed by this plugin.
// Returning or exiting from main will cause the plugin logic to stop, and trigger a safe Shutdown with Stop
func Main(Client *client.SimpleClient, args ...interface{}) {

	// Main plugin loop.
	for {
		// ...

		time.Sleep(DefaultPluginCycleTime)
	}
}

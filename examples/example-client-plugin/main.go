package main

import (
	"github.com/Bearnie-H/easy-tls/client"
	"time"
)

// Main is the top-level ACTION performed by this plugin.  Returning or exiting from main is equivalent to stopping the plugin.
func Main(Client *client.SimpleClient, args ...interface{}) {

	// Main plugin loop.
	for {
		// ...

		time.Sleep(DefaultPluginCycleTime)
	}
}

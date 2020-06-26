package main

import (
	"time"

	"github.com/Bearnie-H/easy-tls/client"
)

// Main is the top-level ACTION performed by this plugin.
func Main(Client *client.SimpleClient, args ...interface{}) {

	// Main plugin loop.
	for {
		// ...

		time.Sleep(DefaultPluginCycleTime)
	}
}

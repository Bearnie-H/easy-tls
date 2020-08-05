package main

import (
	"time"

	"github.com/Bearnie-H/easy-tls/client"
)

// Main is the top-level ACTION performed by this plugin.
func Main(Client *client.SimpleClient, args ...interface{}) {

	// Create a top-level cancellable context to oversee the entire main function.
	ctx, Key := Contexts.NewContext()
	defer Contexts.RemoveContext(Key)

	// Set up a timer to let the module perform whatever action at some regular interval
	t := time.NewTicker(DefaultPluginCycleTime)
	defer t.Stop()

	// Main plugin loop
	// If the context has been cancelled by an external call to Stop(),
	// simply abort and return, otherwise block until it's time to do whatever
	// this module does.
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			// ...
		}
	}
}

package main

import (
	"time"

	"github.com/Bearnie-H/easy-tls/client"
)

// Main is the top-level ACTION performed by this plugin.
func Main(Client *client.SimpleClient, args ...interface{}) {

	tf := time.NewTicker(time.Second)
	ts := time.NewTicker(DefaultPluginCycleTime)

	defer tf.Stop()
	defer ts.Stop()

	// Main plugin loop.
	for !Killed.Load().(bool) {
		select {
		case <-tf.C:
			continue
		case <-ts.C:
			// ...
		}
	}
}

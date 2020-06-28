package main

import (
	"sync/atomic"
)

// PluginType tells which type of plugin this is, server or client.
const PluginType string = "server"

// Killed represents whether or not the plugin has been killed/stopped.
var Killed atomic.Value

// Must be present but empty.
func main() {}

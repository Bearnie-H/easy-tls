package main

import (
	"sync/atomic"

	"github.com/Bearnie-H/easy-tls/plugins"
)

// StatusChannel represents the channel this plugin can use to output its status messages.
var StatusChannel chan plugins.PluginStatus = nil

// Killed represents whether or not the plugin has been killed/stopped.
var Killed atomic.Value

// Must be present but empty.
func main() {}

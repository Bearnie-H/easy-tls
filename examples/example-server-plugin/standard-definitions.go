package main

import (
	"sync/atomic"
	"time"

	"github.com/Bearnie-H/easy-tls/plugins"
)

// StatusChannel represents the channel this plugin can use to output its status messages.
var StatusChannel chan plugins.PluginStatus = nil

// Killed represents whether or not the plugin has been killed/stopped.
var Killed atomic.Value

// DefaultPluginCycleTime represents the amount of time to wait between writing out the running/killed status message.
var DefaultPluginCycleTime time.Duration = time.Minute * 5

// Must be present but empty.
func main() {}

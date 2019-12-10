package main

import (
	"sync/atomic"
	"time"

	"github.com/Bearnie-H/easy-tls/plugins"
)

// StatusChannel represents the channel this plugin can use to output its status messages.
var StatusChannel chan plugins.PluginStatus = nil

// Killed defines whether or not this plugin has been signalled to be killed/stopped
var Killed *atomic.Value

// DefaultPluginCycleTime represents the amount of time to wait between cycles of the plugin Main loop.
var DefaultPluginCycleTime time.Duration = time.Second * 15

// Must be present but empty.
func main() {}
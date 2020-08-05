package main

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/Bearnie-H/easy-tls/plugins"
)

// Killed defines whether or not this plugin has been signalled to be killed/stopped
var Killed atomic.Value

// ThreadCount is the plugin-global variable used to assert all spawned go-routines
// of the plugin are stopped before Stop() can return control to the plugin
// agent managing this module.
//
// This should be incremented and defer decremented by all go-routines with timing
// under control of this module.
var ThreadCount = &sync.WaitGroup{}

// Contexts tracks all active contexts for the plugin, allowing them to be safely cancelled
// by the Stop() function.
var Contexts *plugins.ContextManager = nil

// PluginType tells which type of plugin this is, server, client or generic.
const PluginType string = "client"

// DefaultPluginCycleTime represents the amount of time to wait between cycles of the plugin Main loop.
var DefaultPluginCycleTime time.Duration = time.Second * 15

// Must be present but empty.
func main() {}

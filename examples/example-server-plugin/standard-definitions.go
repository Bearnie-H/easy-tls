package main

import (
	"sync"
	"sync/atomic"

	"github.com/Bearnie-H/easy-tls/plugins"
)

// Killed defines whether or not this plugin has been signalled to be killed/stopped
var Killed atomic.Value

// Registered answers: "Have the routes of this module already been returned?"
// If so, don't return them again on future Init() calls to prevent multiple
// registration of the handlers with the server.
var Registered sync.Once

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
const PluginType string = "server"

// Must be present but empty.
func main() {}

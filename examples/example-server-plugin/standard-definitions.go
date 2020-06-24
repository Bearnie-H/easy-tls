package main

import (
	"sync"
	"sync/atomic"

	"github.com/Bearnie-H/easy-tls/plugins"
)

// StatusChannel represents the channel this plugin can use to output its status messages.
var StatusChannel chan plugins.PluginStatus = nil

// StatusLock will lock the status channel, protecting it from data races
// between the various writing go-routines and the Stop function
var StatusLock *sync.Mutex = &sync.Mutex{}

// Killed represents whether or not the plugin has been killed/stopped.
var Killed atomic.Value

// Must be present but empty.
func main() {}

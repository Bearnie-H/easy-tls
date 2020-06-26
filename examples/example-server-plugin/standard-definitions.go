package main

import (
	"sync/atomic"
)

// Killed represents whether or not the plugin has been killed/stopped.
var Killed atomic.Value

// Must be present but empty.
func main() {}

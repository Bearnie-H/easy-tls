package main

import (
	"fmt"
	"sync"

	"github.com/Bearnie-H/easy-tls/plugins"
)

// StatusChannelBufferLength defines how large to buffer the Status Channel for.  This should be large enough to allow multiple go-routines to read and write the channel without blocking overly, but also not take up undue memory.
const StatusChannelBufferLength int = 10

// StatusChannel is the encapsulated writer to use when writing messages out to the world.
// This handles synchronization, opening, closing, and locking.
var StatusChannel *StatusWriter = &StatusWriter{
	out:  make(chan plugins.PluginStatus, StatusChannelBufferLength),
	lock: &sync.Mutex{},
}

// StatusWriter is the encapsulation of how a plugin writes status messages out to the world.
type StatusWriter struct {
	out  chan plugins.PluginStatus
	lock *sync.Mutex
}

// Channel will return the internal channel used by the plugin, to export it to the world
func (W *StatusWriter) Channel() (<-chan plugins.PluginStatus, error) {
	return W.out, nil
}

// Printf will format and print out a status message usng the given Message fmt string and optional error.
func (W *StatusWriter) Printf(Message string, Error error, args ...interface{}) {
	W.lock.Lock()
	defer W.lock.Unlock()

	W.out <- plugins.PluginStatus{
		Message: fmt.Sprintf("[%s]: %s", PluginName, fmt.Sprintf(Message, args...)),
		Error:   Error,
		IsFatal: false,
	}
}

// Fatalf will format and print out a status message usng the given Message fmt string
// and optional error. This will tell the framework to hard stop the plugin.
func (W *StatusWriter) Fatalf(Message string, Error error, args ...interface{}) {
	W.out <- plugins.PluginStatus{
		Message: fmt.Sprintf("[%s]: %s", PluginName, fmt.Sprintf(Message, args...)),
		Error:   Error,
		IsFatal: true,
	}
}

// Close will safely close and lock the channel, preventing all other access.
func (W *StatusWriter) Close(err error) {
	W.Printf("Stopped module %s", err, PluginName)
	W.lock.Lock()
	close(W.out)
}

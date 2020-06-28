package plugins

import (
	"fmt"
	"sync"
)

// StatusWriter is the encapsulation of how a plugin writes status messages out to the world.
type StatusWriter struct {
	out  chan PluginStatus
	lock *sync.Mutex
	name string
}

// OpenStatusWriter will open a new StatusWriter for a given plugin.
func OpenStatusWriter(Length int, Name string) *StatusWriter {
	if Length == 0 {
		Length = 1
	}
	return &StatusWriter{
		out:  make(chan PluginStatus, Length),
		lock: &sync.Mutex{},
		name: Name,
	}
}

// Channel will return the internal channel used by the plugin, to export it to the world
func (W *StatusWriter) Channel() (<-chan PluginStatus, error) {
	return W.out, nil
}

// Printf will format and print out a status message usng the given Message fmt string and optional error.
func (W *StatusWriter) Printf(Message string, Error error, args ...interface{}) {
	W.lock.Lock()
	defer W.lock.Unlock()

	W.out <- PluginStatus{
		Message: fmt.Sprintf("[%s]: %s", W.name, fmt.Sprintf(Message, args...)),
		Error:   Error,
		IsFatal: false,
	}
}

// Fatalf will format and print out a status message usng the given Message fmt string
// and optional error. This will tell the framework to hard stop the plugin.
func (W *StatusWriter) Fatalf(Message string, Error error, args ...interface{}) {
	W.out <- PluginStatus{
		Message: fmt.Sprintf("[%s]: %s", W.name, fmt.Sprintf(Message, args...)),
		Error:   Error,
		IsFatal: true,
	}
}

// Out allows sending a pre-formatted status message out
func (W *StatusWriter) Out(S PluginStatus) {
	W.lock.Lock()
	defer W.lock.Unlock()
	W.out <- S
}

// Close will safely close and lock the channel, preventing all other access.
func (W *StatusWriter) Close(err error) {
	W.Printf("Stopped module %s", err, W.name)
	W.lock.Lock()
	close(W.out)
}

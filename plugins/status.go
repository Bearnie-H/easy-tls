package plugins

import (
	"fmt"
	"sync"
)

// PluginStatus represents a single status message from a given EasyTLS-compliant plugin.
type PluginStatus struct {
	message string
	err     error
	fatal   bool
}

// NewPluginStatus creates and returns a new status message, able to be sent via Statuswriter
func NewPluginStatus(Message string, Err error, Fatal bool) PluginStatus {
	return PluginStatus{
		message: Message,
		err:     Err,
		fatal:   Fatal,
	}
}

func (S *PluginStatus) String() string {
	switch {
	case S.fatal:
		return fmt.Sprintf("FATAL ERROR: %s - %s", S.message, S.err)
	case S.err != nil:
		return fmt.Sprintf("Warning: %s - %s", S.message, S.err)
	default:
		return S.message
	}
}

func (S *PluginStatus) Error() string {
	switch {
	case S.fatal:
		return fmt.Sprintf("FATAL ERROR: %s - %s", S.message, S.err)
	case S.err != nil:
		return fmt.Sprintf("Warning: %s - %s", S.message, S.err)
	default:
		return ""
	}
}

// StatusWriter is the encapsulation of how a plugin writes status messages out to the world.
type StatusWriter struct {
	out    chan PluginStatus
	lock   *sync.Mutex
	active bool
	name   string
	length int
}

// OpenStatusWriter will open a new StatusWriter for a given plugin.
func OpenStatusWriter(Length int, Name string) *StatusWriter {

	if Length == 0 {
		Length = 1
	}

	return &StatusWriter{
		out:    make(chan PluginStatus, Length),
		lock:   &sync.Mutex{},
		active: true,
		name:   Name,
		length: Length,
	}
}

// Channel will return the internal channel used by the plugin, to export it to the world
func (W *StatusWriter) Channel() (<-chan PluginStatus, error) {

	W.lock.Lock()
	defer W.lock.Unlock()

	if !W.active || W.out == nil {
		W.out = make(chan PluginStatus, W.length)
		W.active = true
	}

	return W.out, nil
}

// Printf will format and print out a status message usng the given Message fmt string and optional error.
func (W *StatusWriter) Printf(Message string, Err error, args ...interface{}) {

	W.lock.Lock()
	defer W.lock.Unlock()

	if !W.active {
		return
	}

	W.out <- NewPluginStatus(
		fmt.Sprintf("[%s]: %s", W.name, fmt.Sprintf(Message, args...)),
		Err,
		false,
	)
}

// Fatalf will format and print out a status message usng the given Message fmt string
// and optional error. This will tell the framework to hard stop the plugin.
func (W *StatusWriter) Fatalf(Message string, Err error, args ...interface{}) {

	W.lock.Lock()
	defer W.lock.Unlock()

	if !W.active {
		return
	}

	W.out <- NewPluginStatus(
		fmt.Sprintf("[%s]: %s", W.name, fmt.Sprintf(Message, args...)),
		Err,
		true,
	)
}

// Out allows sending a pre-formatted status message out
func (W *StatusWriter) Out(S PluginStatus) {

	W.lock.Lock()
	defer W.lock.Unlock()

	if !W.active {
		return
	}

	W.out <- NewPluginStatus(
		fmt.Sprintf("[%s]: %s", W.name, S.message),
		S.err,
		S.fatal,
	)
}

// Close will safely close and lock the channel, preventing all other access.
func (W *StatusWriter) Close(err error) {
	W.lock.Lock()
	W.active = false
	defer W.lock.Unlock()
	close(W.out)
	W.out = nil
}

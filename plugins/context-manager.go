package plugins

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"
)

// ErrNoContext indicates there is no context for the provided key.
var ErrNoContext error = errors.New("context manager error: No CancelFunc for key")

// ContextManager represents an abstract object to help manage the
// contexts held/opened by a given module. This object allows for
// the Stop() function to properly and safely Cancel the contexts
// to allow the module to shut down safely.
type ContextManager struct {
	*sync.Mutex
	active map[int64]context.CancelFunc
	r      rand.Source
	closed bool
}

// NewContextManager will initialize a new ContextManager, making it ready to use.
func NewContextManager() *ContextManager {
	return &ContextManager{
		Mutex:  &sync.Mutex{},
		active: make(map[int64]context.CancelFunc),
		r:      rand.NewSource(time.Now().UnixNano()),
		closed: false,
	}
}

// NewContext creates and new context.WithCancel() context to be used by the caller.
// This function saves the CancelFunc for the generated context, returning
// and int64 token which can be used to manipulate the CancelFunc later.
//
// This returned context should be extended to any other type of context as necessary,
// such as WithDeadline() or WithValue(). This simply ensures the parent context is cancellable
// to prevent undesirable waits when stopping plugins.
func (C *ContextManager) NewContext() (ctx context.Context, x int64) {

	C.Lock()
	defer C.Unlock()

	// Create a new context which is able to be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Generate a unique token for this context
	for x = C.r.Int63(); C.active[x] != nil; x = C.r.Int63() {
	}

	// If the manager is closed, don't internalize the context and pre-cancel it
	// so anything that uses it will only see the already cancelled context
	if C.closed {
		cancel()
		return ctx, x
	}

	// Internalize the returned CancelFunc
	C.active[x] = cancel

	// return the context as well as the key to use to manipulate it later.
	return ctx, x
}

// RemoveContext will remove the CancelFunc for a given key, indicating that
// the related context has successfully finished whatever task it was related to
// and no longer needs to be able to be cancelled externally.
//
// Standard operation involves calls like:
//
//	ctx, Key := Manager.NewContext()
//	defer Manager.RemoveContext(Key)
//	// functionality here to use ctx.
//
// Or:
//
//	ctx, Key := Manager.NewContext()
//	// Use ctx
//	Manager.RemoveContext(Key)
func (C *ContextManager) RemoveContext(Key int64) {
	C.Lock()
	defer C.Unlock()

	if cancel, exist := C.active[Key]; exist {
		cancel()
	}

	delete(C.active, Key)
}

// GetCancel will return the CancelFunc for a given Key, allowing
// the caller to act on it.
func (C *ContextManager) GetCancel(Key int64) (context.CancelFunc, error) {

	C.Lock()
	defer C.Unlock()

	Cancel := C.active[Key]

	if Cancel == nil {
		return nil, ErrNoContext
	}

	return Cancel, nil
}

// Cancel will call the CancelFunc (if it exists) for the given Key,
// removing the stored function from the manager once the cancel completes.
func (C *ContextManager) Cancel(Key int64) error {

	C.Lock()
	defer C.Unlock()

	Cancel := C.active[Key]

	if Cancel == nil {
		return ErrNoContext
	}

	Cancel()
	delete(C.active, Key)
	return nil
}

// Close will close a ContextManager, cancelling all contexts which have not
// been removed from the manager yet.
func (C *ContextManager) Close() {

	C.Lock()
	C.closed = true
	defer C.Unlock()

	// Call the CancelFunc's for all active contexts
	for _, Cancel := range C.active {
		Cancel()
	}

	// Clear the map
	C.active = make(map[int64]context.CancelFunc)
}

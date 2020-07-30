package main

import (
	"fmt"
)

// Init is the function to start the plugin logic.
func Init(args ...interface{}) error {
	defer func() {
		r := recover()
		if e, ok := r.(error); ok {
			StatusChannel.Fatalf("easytls plugin error: Panic during execution", e)
		}
	}()

	ThreadCount.Add(1)
	defer ThreadCount.Done()

	// Perform the non-specific module initialization steps.
	if err := defaultInitialization(args...); err != nil {
		return fmt.Errorf("easytls module error: Failed to perform standard initialization - %s", err)
	}

	// Perform the module-specific initialization steps.
	if err := moduleInitialization(args...); err != nil {
		return fmt.Errorf("easytls module error: Failed to perform module-specific initialization - %s", err)
	}

	// Now the all of the initialization steps are finished, spawn a go-routine to implement the "Main" logic of this plugin.
	go func(args ...interface{}) {
		defer Stop()
		ThreadCount.Add(1)
		defer ThreadCount.Done()
		Main(args...)
	}(args...)

	return nil
}

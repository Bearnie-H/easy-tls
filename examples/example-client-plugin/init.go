package main

import (
	"fmt"

	"github.com/Bearnie-H/easy-tls/client"
)

// Init is the function to start the plugin logic.
func Init(Client *client.SimpleClient, args ...interface{}) error {

	// Defer a call to recover, to assert that no panic can possibly
	// cross the API boundary between a plugin and the overarching framework.
	defer func() {
		r := recover()
		if e, ok := r.(error); ok {
			StatusChannel.Fatalf("easytls plugin error: Panic during execution", e)
		}
	}()

	ThreadCount.Add(1)
	defer ThreadCount.Done()

	// Perform the non-specific module initialization steps.
	if err := defaultInitialization(Client, args...); err != nil {
		return fmt.Errorf("easytls module error: Failed to perform standard initialization - %s", err)
	}

	// Perform the module-specific initialization steps.
	if err := moduleInitialization(Client, args...); err != nil {
		return fmt.Errorf("easytls module error: Failed to perform module-specific initialization - %s", err)
	}

	// Now the all of the initialization steps are finished, spawn a go-routine to implement the "Main" logic of this plugin.
	ThreadCount.Add(1)
	go func(Client *client.SimpleClient, args ...interface{}) {
		defer ThreadCount.Done()
		Main(Client, args...)
	}(Client, args...)

	return nil
}

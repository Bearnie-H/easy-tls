package main

import (
	"fmt"
	"time"

	"github.com/Bearnie-H/easy-tls/client"
)

// Init is the function to start the plugin logic.
func Init(Client *client.SimpleClient, args ...interface{}) error {

	// Perform the non-specific module initialization steps.
	if err := defaultInitialization(Client, args...); err != nil {
		return fmt.Errorf("easytls module error: Failed to perform standard initialization - %s", err)
	}

	// Perform the module-specific initialization steps.
	if err := moduleInitialization(Client, args...); err != nil {
		return fmt.Errorf("easytls module error: Failed to perform module-specific initialization - %s", err)
	}

	// Now the all of the initialization steps are finished, spawn a go-routine to implement the "Main" logic of this plugin.
	go func(Client *client.SimpleClient, args ...interface{}) {
		Main(Client, args...)
	}(Client, args...)

	// Spawn a go-routine which tracks the module's running state, regardless of any logging present within the module.
	// This is just a sanity check, so logging can otherwise be deferred to success and error states, not necessary begin->fail/succeed.
	go func() {
		for {
			if !Killed.Load().(bool) {
				WriteStatus("Module [ %s ] running", nil, false, Name())
				time.Sleep(DefaultPluginCycleTime)
			} else {
				WriteStatus("Module [ %s ] killed!", nil, false, Name())
				break
			}
		}
	}()

	return nil
}

package main

import (
	"fmt"

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

	go func(Client *client.SimpleClient, args ...interface{}) {
		Main(Client, args...)
	}(Client, args...)

	return nil
}

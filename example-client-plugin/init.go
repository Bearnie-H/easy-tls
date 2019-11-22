package main

import (
	"fmt"

	easytls "github.com/Bearnie-H/easy-tls"
)

// Init is the function to start the plugin logic.
func Init(Client *easytls.SimpleClient) error {

	// Perform the non-specific module initialization steps.
	if err := defaultInitialization(Client); err != nil {
		return fmt.Errorf("easytls module error: Failed to perform standard initialization - %s", err)
	}

	// Perform the module-specific initialization steps.
	if err := moduleInitialization(Client); err != nil {
		return fmt.Errorf("easytls module error: Failed to perform module-specific initialization - %s", err)
	}

	go func(Client *easytls.SimpleClient) {
		Main(Client)
	}(Client)

	return nil
}

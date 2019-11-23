package main

import (
	"errors"
	"fmt"

	easytls "github.com/Bearnie-H/easy-tls"
)

func defaultInitialization() error {

	// ...

	return nil
}

// Status will prepare the StatusChannel and return it, a Non-Nil error implies a failure and means the channel is NOT initialized.
func Status() (<-chan easytls.PluginStatus, error) {

	// Make this call idempotent.
	if StatusChannel != nil {
		return StatusChannel, nil
	}

	// Create a single non-blocking channel
	StatusChannel = make(chan easytls.PluginStatus, 1)

	return StatusChannel, nil
}

// Version will compare the version of the Framework with what this module is defined to be compatable with.
func Version(FrameworkVersion easytls.SemanticVersion) error {
	if easytls.Accepts(FrameworkVersion, RequiresFrameworkMinVersion, RequiresFrameworkMaxVersion) {
		return nil
	}
	return fmt.Errorf("easytls module error: Incompatable versions - %s !<= %s !<= %s", RequiresFrameworkMinVersion.String(), FrameworkVersion.String(), RequiresFrameworkMaxVersion.String())
}

// Name will return the name of this module, in a canonical format.
func Name() (string, error) {
	return fmt.Sprintf("%s-%s", PluginName, PluginVersion.String()), nil
}

// WriteStatus is the standard mechanism for writing a status message out to the framework.  This function can and should be passed in to sub-packages as necessary within the plugin, along with the StatusChannel itself (or at least a pointer to these).
func WriteStatus(NewStatus easytls.PluginStatus) error {

	// Cannot write a status to an uninitialized channel
	if StatusChannel == nil {
		return errors.New("easytls module error: StatusChannel not initialized")
	}

	// If the channel is holding a message, clear it in a non-blocking way
	select {
	case <-StatusChannel:
	default:
	}

	// Send the new status message.
	StatusChannel <- NewStatus

	return nil
}
package main

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/Bearnie-H/easy-tls/client"
	"github.com/Bearnie-H/easy-tls/plugins"
)

// StatusChannelBufferLength defines how large to buffer the Status Channel for.  This should be large enough to allow multiple go-routines to read and write the channel without blocking overly, but also not take up undue memory.
const StatusChannelBufferLength int = 10

func defaultInitialization(Client *client.SimpleClient, args ...interface{}) error {

	Killed.Store(false)

	// ...

	return nil
}

// Status will prepare the StatusChannel and return it, a Non-Nil error implies a failure and means the channel is NOT initialized.
func Status() (<-chan plugins.PluginStatus, error) {

	// Make this call idempotent.
	if StatusChannel != nil {
		return StatusChannel, nil
	}

	// Create a single non-blocking channel
	StatusChannel = make(chan plugins.PluginStatus, StatusChannelBufferLength)

	return StatusChannel, nil
}

// Version will compare the version of the Framework with what this module is defined to be compatable with.
func Version(FrameworkVersion plugins.SemanticVersion) error {
	if plugins.Accepts(FrameworkVersion, RequiresFrameworkMinVersion, RequiresFrameworkMaxVersion) {
		return nil
	}
	return fmt.Errorf("easytls module error: Incompatable versions - %s !<= %s !<= %s", RequiresFrameworkMinVersion.String(), FrameworkVersion.String(), RequiresFrameworkMaxVersion.String())
}

// Name will return the name of this module, in a canonical format.
func Name() string {
	return fmt.Sprintf("%s-%s (%s)", PluginName, PluginVersion.String(), PluginType)
}

// WriteStatus is the standard mechanism for writing a status message out to the framework.  This function can and should be passed in to sub-packages as necessary within the plugin, along with the StatusChannel itself (or at least a pointer to these).
func WriteStatus(Message string, Error error, Fatal bool, args ...interface{}) error {

	NewStatus := plugins.PluginStatus{
		Message: fmt.Sprintf("[%s]: %s", PluginName, fmt.Sprintf(Message, args...)),
		Error:   Error,
		IsFatal: Fatal,
	}

	// Cannot write a status to an uninitialized channel
	if StatusChannel == nil {
		return errors.New("easytls module error: StatusChannel not initialized")
	}

	if Killed.Load().(bool) {
		return errors.New("easytls module error: Cannot send over StatusChannel after module is Killed")
	}

	// Send the new status message.
	StatusChannel <- NewStatus

	return nil
}

func getFolderBase() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	return path.Join(path.Dir(ex), PluginName), nil
}

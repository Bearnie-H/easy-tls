package main

import (
	"fmt"
	"os"
	"path"

	"github.com/Bearnie-H/easy-tls/plugins"
)

func defaultInitialization(args ...interface{}) error {

	Killed.Store(false)

	// ...

	return nil
}

// Status will prepare the StatusChannel and return it, a Non-Nil error implies a failure and means the channel is NOT initialized.
func Status() (<-chan plugins.PluginStatus, error) {
	return StatusChannel.Channel()
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

func getFolderBase() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	return path.Join(path.Dir(ex), PluginName), nil
}

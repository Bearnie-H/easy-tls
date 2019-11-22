package main

import (
	"fmt"

	easytls "github.com/Bearnie-H/easy-tls"
)

/*
	You should not need to edit this file unless the base-line actions of the Client Plugin are substantially different than standard.
*/

func defaultInitialization(Client *easytls.SimpleClient) error {

	// ...

	return nil
}

// Status will prepare the StatusChannel and return it, a Non-Nil error implies a failure and means the channel is NOT initialized.
func Status() (chan easytls.PluginStatus, error) {
	if StatusChannel != nil {
		return StatusChannel, nil
	}

	return nil, nil
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

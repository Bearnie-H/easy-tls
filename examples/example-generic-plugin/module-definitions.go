package main

import (
	"github.com/Bearnie-H/easy-tls/plugins"
)

// PluginName is the name of the current plugin.
const PluginName string = "DEFINE_ME"

// Semantic Versioning Information
var (
	// What minimum framework version is supported/required
	RequiresFrameworkMinVersion = plugins.SemanticVersion{
		MajorRelease: 1,
		MinorRelease: 4,
		Build:        1,
	}

	// What maximum framework version is supported/required
	RequiresFrameworkMaxVersion = plugins.SemanticVersion{
		MajorRelease: 2,
		MinorRelease: 1,
		Build:        1,
	}
)

func moduleInitialization(args ...interface{}) error {

	// ... Your module-specific initialization steps here

	return nil
}

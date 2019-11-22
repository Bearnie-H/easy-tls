package main

import (
	"io"
	"os"

	easytls "github.com/Bearnie-H/easy-tls"
)

// PluginName is the name of the current plugin.
const PluginName string = "DEFINE_ME"

// LogFile represents how the module should log.  This must be WriteCloser, or be provided a NopClose method.
var LogFile io.WriteCloser = os.Stdout

// Semantic Versioning Information
var (
	// What minimum framework version is supported/required
	RequiresFrameworkMinVersion = easytls.SemanticVersion{
		MajorRelease: 1,
		MinorRelease: 1,
		Build:        1,
	}

	// What maximum framework version is supported/required
	RequiresFrameworkMaxVersion = easytls.SemanticVersion{
		MajorRelease: 2,
		MinorRelease: 1,
		Build:        1,
	}

	// What minimum server-side plugin version is supported/required
	RequiresServerPluginMinVersion = easytls.SemanticVersion{
		MajorRelease: 1,
		MinorRelease: 1,
		Build:        1,
	}

	// What maximum server-side plugin version is supported/required
	RequiresServerPluginMaxVersion = easytls.SemanticVersion{
		MajorRelease: 2,
		MinorRelease: 1,
		Build:        1,
	}
)

func moduleInitialization() error {

	// ... Your module-specific initialization steps here

	return nil
}

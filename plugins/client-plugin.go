package plugins

import (
	"path"
	"plugin"

	"github.com/Bearnie-H/easy-tls/client"
)

// ClientPlugin represents a EasyTLS-compatible Plugin to be used with an EasyTLS SimpleClient.
type ClientPlugin struct {

	// A Client Plugin is composed of a generic plugin plus an API Contract.
	Plugin

	// A ClientPlugin must implement the full API Contract.
	ClientPluginAPI
}

// ClientPluginAPI represents the API contract a Client-Plugin must satisfy to be used by this framework.
type ClientPluginAPI struct {

	// Start a plugin.
	//
	// This will provide a SimpleClient object for the Plugin to use for any HTTP(S) operations it should take.
	// If a non-nil error is returned, this indicates that the initialization failed, and the Stop command should be used.
	// No Plugins should function if Init returns a non-nil error.
	Init func(*client.SimpleClient, ...interface{}) error
}

// InitializeClientPlugin will initialize and return a Client Plugin, ready to be registered by a Client Plugin Agent.
func InitializeClientPlugin(Filename string, FrameworkVersion SemanticVersion) (*ClientPlugin, error) {

	// Create the starting plugin object
	P := &ClientPlugin{
		Plugin: Plugin{
			Filename: path.Base(Filename),
			Filepath: path.Dir(Filename),
		},
	}

	// Load the default symbols, erroring out on any failure.
	defaultAPI, err := loadDefaultPluginSymbols(Filename)
	if err != nil {
		return nil, err
	}
	P.Plugin.PluginAPI = defaultAPI

	// Load the client-specific symbols, erroring out on any failure.
	clientAPI, err := loadClientPluginSymbols(Filename)
	if err != nil {
		return nil, err
	}
	P.ClientPluginAPI = clientAPI

	// Assert that the versioning is compatable.
	if err := P.Version(FrameworkVersion); err != nil {
		return nil, err
	}

	return P, nil
}

func loadClientPluginSymbols(Filename string) (ClientPluginAPI, error) {
	API := ClientPluginAPI{}

	rawPlug, err := plugin.Open(Filename)
	if err != nil {
		return API, err
	}

	sym, err := rawPlug.Lookup("Init")
	if err != nil {
		return API, err
	}

	initSym, ok := sym.(func(*client.SimpleClient, ...interface{}) error)
	if !ok {
		return API, err
	}
	API.Init = initSym

	return API, nil
}

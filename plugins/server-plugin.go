package plugins

import (
	"log"
	"plugin"

	"github.com/Bearnie-H/easy-tls/server"
)

type (
	// ServerInitFunc is the defined type of the Init function exported by all compliant server plugins.
	ServerInitFunc = func(...interface{}) ([]server.SimpleHandler, error)
)

// ServerPlugin represents a EasyTLS-compatible Plugin to be used with an EasyTLS SimpleServer.
type ServerPlugin struct {

	// A Server Plugin is composed of a generic plugin plus an API Contract.
	Plugin

	// A ServerPlugin must implement the full API Contract.
	ServerPluginAPI
}

// ServerPluginAPI represents the API contract a Server-Plugin must satisfy to be used by this framework.
type ServerPluginAPI struct {

	// Start a plugin.
	//
	// This will initialize the plugin, and return the set of Routes it can provide back to the SimpleServer.
	Init func(...interface{}) ([]server.SimpleHandler, error)
}

// InitializeServerPlugin will initialize and return a Server Plugin, ready to be registered by a Server Plugin Agent.
func InitializeServerPlugin(Filename string, FrameworkVersion SemanticVersion, Logger *log.Logger) (*ServerPlugin, error) {

	// Create the starting plugin object
	P := &ServerPlugin{
		Plugin: NewPlugin(Filename, Logger),
	}

	// Load the default symbols, erroring out on any failure.
	defaultAPI, err := loadDefaultPluginSymbols(Filename)
	if err != nil {
		return nil, err
	}
	P.Plugin.PluginAPI = defaultAPI

	// Load the client-specific symbols, erroring out on any failure.
	serverAPI, err := loadServerPluginSymbols(Filename)
	if err != nil {
		return nil, err
	}
	P.ServerPluginAPI = serverAPI

	// Assert that the versioning is compatable.
	if err := P.Version(FrameworkVersion); err != nil {
		return nil, err
	}

	return P, nil
}

func loadServerPluginSymbols(Filename string) (ServerPluginAPI, error) {
	API := ServerPluginAPI{}

	rawPlug, err := plugin.Open(Filename)
	if err != nil {
		return API, err
	}

	sym, err := rawPlug.Lookup("Init")
	if err != nil {
		return API, err
	}

	initSym, ok := sym.(ServerInitFunc)
	if !ok {
		return API, err
	}
	API.Init = initSym

	return API, nil
}

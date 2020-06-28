package plugins

import (
	"fmt"
	"plugin"

	"github.com/Bearnie-H/easy-tls/client"
)

type (
	// ClientInitFunc is the defined type of the Init function exported by all compliant client plugins.
	ClientInitFunc = func(*client.SimpleClient, ...interface{}) error
)

// ClientPlugin represents a EasyTLS-compatible Plugin to be used with an EasyTLS SimpleClient.
type ClientPlugin struct {

	// A Client Plugin is composed of a generic plugin plus an API Contract.
	Plugin

	// A ClientPlugin must implement the full API Contract.
	ClientPluginAPI

	// The HTTP Client to pass in to the plugin
	Client *client.SimpleClient
}

// ClientPluginAPI represents the API contract a Client-Plugin must satisfy to be used by this framework.
type ClientPluginAPI struct {

	// Start a plugin.
	//
	// This will provide a SimpleClient object for the Plugin to use for any HTTP(S) operations it should take.
	// If a non-nil error is returned, this indicates that the initialization failed, and the Stop command should be used.
	// No Plugins should function if Init returns a non-nil error.
	Init ClientInitFunc
}

// InitializeClientPlugin will initialize and return a Client Plugin, ready to be registered by a Client Plugin Agent.
func InitializeClientPlugin(Filename string, FrameworkVersion SemanticVersion, Client *client.SimpleClient) (*ClientPlugin, error) {

	// Create the starting plugin object
	P := &ClientPlugin{
		Plugin: NewPlugin(Filename, Client.Logger()),
		Client: Client,
	}

	// Load the default symbols, erroring out on any failure.
	if err := P.loadDefaultPluginSymbols(Filename); err != nil {
		return nil, err
	}

	// Load the client-specific symbols, erroring out on any failure.
	if err := P.loadClientPluginSymbols(Filename); err != nil {
		return nil, err
	}

	// Assert that the versioning is compatable.
	if err := P.Version(FrameworkVersion); err != nil {
		return nil, err
	}

	return P, nil
}

// Start will start a given Client Plugin, setting up status reading and the necessary synchronization for stopping safely.
func (P *ClientPlugin) Start() {

	Name := P.Name()

	// Set up a catch if Init() panics
	defer func(P *ClientPlugin) {
		if r := recover(); r != nil {
			P.Logger.Printf("easytls plugin error: Plugin [ %s ] panic during Init()!", Name)
			if err := P.Kill(); err != nil {
				P.Logger.Printf("easytls plugin error: Plugin [ %s ] errored while stopping after Init() panic - %s", Name, err)
			}
		}
	}(P)

	if err := P.Init(P.Client, P.InputArguments...); err != nil {
		P.Logger.Printf("easytls plugin error: Plugin [ %s ] failed to Init() - %s", Name, err)
		if err := P.Kill(); err != nil {
			P.Logger.Printf("easytls plugin error: Plugin [ %s ] errored while stopping after Init() failure - %s", Name, err)
		}
	}
}

func (P *ClientPlugin) loadClientPluginSymbols(Filename string) error {

	rawPlug, err := plugin.Open(Filename)
	if err != nil {
		return err
	}

	SymbolName := "Init"
	sym, err := rawPlug.Lookup(SymbolName)
	if err != nil {
		return err
	}

	switch sym.(type) {
	case ClientInitFunc:
		P.Init = sym.(ClientInitFunc)
	default:
		return fmt.Errorf("easytls plugin error: Invalid %s() signature, expected [ %s ] - got [ %s ]", SymbolName, getFuncSignature(P.Init), getFuncSignature(sym))
	}

	return nil
}

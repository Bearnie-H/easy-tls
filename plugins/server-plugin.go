package plugins

import (
	"fmt"
	"log"
	"plugin"

	"github.com/Bearnie-H/easy-tls/server"
)

type (
	// ServerInitHandlersFunc is the defined type of the Init function exported
	// by compliant plugins to register a set of routes with the default router.
	ServerInitHandlersFunc = func(...interface{}) ([]server.SimpleHandler, error)

	// ServerInitSubrouterFunc is the defined type of the Init function exported
	// by compliant plugins to register a set of routes with a dedicated SubRouter
	// with the returned string acting as the URL PathPrefix.
	ServerInitSubrouterFunc = func(...interface{}) ([]server.SimpleHandler, string, error)
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

	// InitHandlers is the plugin-exported function which provides a flat array
	// of SimpleHandlers to register with the default Router of the server.
	// This is the Init() type to satisfy if a plugin returns a set of Handlers
	// which do not all share a common PathPrefix, and are therefore unsuitable
	// to be sub-routed by URL.
	InitHandlers ServerInitHandlersFunc

	// InitHandlers is the plugin-exported function which provides a flat array
	// of SimpleHandlers to register by creating a dedicated SubRouter with the
	// returned PathPrefix.
	//
	// This is the Init() type to satisfy if a plugin returns a set of Handlers
	// which do all share a common PathPrefix, and are therefore suitable
	// to be sub-routed by URL.
	InitSubrouter ServerInitSubrouterFunc
}

// InitializeServerPlugin will initialize and return a Server Plugin, ready to be registered by a Server Plugin Agent.
func InitializeServerPlugin(Filename string, FrameworkVersion SemanticVersion, Logger *log.Logger) (*ServerPlugin, error) {

	// Create the starting plugin object
	P := &ServerPlugin{
		Plugin: NewPlugin(Filename, Logger),
	}

	// Load the default symbols, erroring out on any failure.
	if err := P.loadDefaultPluginSymbols(Filename); err != nil {
		return nil, err
	}

	// Load the server-specific symbols, erroring out on any failure.
	if err := P.loadServerPluginSymbols(Filename); err != nil {
		return nil, err
	}

	// Assert that the versioning is compatable.
	if err := P.Version(FrameworkVersion); err != nil {
		return nil, err
	}

	return P, nil
}

func (P *ServerPlugin) loadServerPluginSymbols(Filename string) error {

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
	case ServerInitHandlersFunc:
		P.InitHandlers = sym.(ServerInitHandlersFunc)
	case ServerInitSubrouterFunc:
		P.InitSubrouter = sym.(ServerInitSubrouterFunc)
	default:
		return fmt.Errorf("easytls plugin error: Invalid %s() signature, expected [ %s ] or [ %s ] - got [ %s ]", SymbolName, getFuncSignature(P.InitHandlers), getFuncSignature(P.InitSubrouter), getFuncSignature(sym))
	}

	return nil
}

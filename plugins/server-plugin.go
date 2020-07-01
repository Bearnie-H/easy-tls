package plugins

import (
	"fmt"
	"plugin"
	"time"

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

// ServerPlugin extends GenericPlugin to fulfill the Module interface
// with Init() functions specific to server-type plugins.
type ServerPlugin struct {
	GenericPlugin

	// Reference to the controlling Agent
	agent *ServerAgent

	// Init fields
	initHandlers  ServerInitHandlersFunc
	initSubrouter ServerInitSubrouterFunc
}

// Reload will fully reload a module.
// This will stop a running module, load the symbols fresh from disk
// and then start the module again.
func (p *ServerPlugin) Reload() error {

	var err error

	if err = p.Stop(); err != nil {
		p.agent.Logger().Printf("plugin reload error: Error stopping plugin [ %s ] for reload - %s", p.Name(), err)
		return err
	}

	p.unloadSymbols()

	if err = p.Load(); err != nil {
		p.agent.Logger().Printf("plugin reload error: Error loading plugin [ %s ] for reload - %s", p.Name(), err)
		return err
	}

	if err = p.Start(); err != nil {
		p.agent.Logger().Printf("plugin reload error: Error starting plugin [ %s ] for reload - %s", p.Name(), err)
		return err
	}

	return err
}

// Load the symbols of the module
func (p *ServerPlugin) Load() error {

	p.mu.Lock()
	defer p.mu.Unlock()

	switch p.state {
	case stateNotLoaded:
	case stateLoaded:
		p.agent.Logger().Printf("Cannot Load() module [ %s ], symbols already loaded", p.Name())
		return nil
	case stateActive:
		p.agent.Logger().Printf("Cannot Load() module [ %s ], already running", p.Name())
		return nil
	}

	// Load the default symbols.
	if err := p.loadDefaultSymbols(); err != nil {
		p.agent.Logger().Printf("plugin load error: Failed to load default symbols from file [ %s ]", p.filename)
		return err
	}

	// Load the type-specific symbols.
	if err := p.loadServerSymbols(); err != nil {
		p.agent.Logger().Printf("plugin load error: Failed to load server symbols from file [ %s ]", p.filename)
		return err
	}

	p.state = stateLoaded

	p.agent.Logger().Printf("Loaded symbols for module [ %s ]", p.Name())

	return nil
}

// Start will start the module, performing any initialization and putting
// it into a state where the logic included by the plugin can be used.
func (p *ServerPlugin) Start() error {

	p.mu.Lock()
	defer p.mu.Unlock()

	switch p.state {
	case stateNotLoaded:
		if err := p.Load(); err != nil {
			return err
		}
	case stateLoaded:
	case stateActive:
		p.agent.Logger().Printf("Cannot Start() module [ %s ], already running", p.Name())
		return nil
	}

	if err := p.ReadStatus(); err != nil {
		return err
	}

	switch {
	case p.initHandlers != nil:
		{
			routes, err := p.initHandlers(p.args...)
			if err != nil {
				return err
			}
			p.agent.server.AddHandlers(p.agent.router, routes...)
		}
	case p.initSubrouter != nil:
		{
			routes, prefix, err := p.initSubrouter(p.args...)
			if err != nil {
				return err
			}
			p.agent.server.AddSubrouter(p.agent.router, prefix, routes...)
		}
	default:
		return fmt.Errorf("plugin error: No Init() function loaded for module [ %s ]", p.Name())
	}

	p.state = stateActive
	p.started = time.Now()
	p.agent.Logger().Printf("Started module [ %s ]", p.Name())

	return nil
}

func (p *ServerPlugin) loadServerSymbols() error {

	raw, err := plugin.Open(p.filename)
	if err != nil {
		return err
	}

	if err := p.loadServerSymbol(raw, "Init"); err != nil {
		return err
	}

	// Add more symbols here...

	return nil
}

func (p *ServerPlugin) loadServerSymbol(rawPlugin *plugin.Plugin, SymbolName string) error {

	s, err := rawPlugin.Lookup(SymbolName)
	if err != nil {
		return err
	}

	// Dispatch on the type of the returned symbol.
	// Each of the expected or possible symbols must have a corresponding case
	// As more possible symbols are added, simply extend the logic to account
	// for all the types.
	switch s.(type) {
	case ServerInitHandlersFunc:
		p.initHandlers = s.(ServerInitHandlersFunc)
	case ServerInitSubrouterFunc:
		p.initSubrouter = s.(ServerInitSubrouterFunc)
	default:
		return fmt.Errorf("plugin load error: Unknown type returned for symbol [ %s ], got [ %s ]", SymbolName, getFuncSignature(s))
	}

	return nil
}

func (p *ServerPlugin) unloadSymbols() {

	p.mu.Lock()
	defer p.mu.Unlock()

	p.state = stateNotLoaded

	p.unloadDefaultSymbols()

	p.initHandlers = nil
	p.initSubrouter = nil
	// ...
}

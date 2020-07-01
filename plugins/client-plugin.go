package plugins

import (
	"fmt"
	"plugin"
	"time"

	"github.com/Bearnie-H/easy-tls/client"
)

type (
	// ClientInitFunc is the defined type of the Init function exported
	// by compliant plugins to spawn a new Client plugin.
	ClientInitFunc = func(*client.SimpleClient, ...interface{}) error
)

// ClientPlugin extends GenericPlugin to fulfill the Module interface
// with Init() functions specific to client-type plugins.
type ClientPlugin struct {
	GenericPlugin

	// Reference to the controlling Agent
	agent *ClientAgent

	// Init fields
	init ClientInitFunc
}

// Reload will fully reload a module.
// This will stop a running module, load the symbols fresh from disk
// and then start the module again.
func (p *ClientPlugin) Reload() error {

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
func (p *ClientPlugin) Load() error {

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
	if err := p.loadClientSymbols(); err != nil {
		p.agent.Logger().Printf("plugin load error: Failed to load client symbols from file [ %s ]", p.filename)
		return err
	}

	p.state = stateLoaded

	p.agent.Logger().Printf("Loaded symbols for module [ %s ]", p.Name())

	return nil
}

// Start will start the module, performing any initialization and putting
// it into a state where the logic included by the plugin can be used.
func (p *ClientPlugin) Start() error {

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
	case p.init != nil:
		{
			if err := p.init(p.agent.client, p.args...); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("plugin error: No Init() function loaded for module [ %s ]", p.Name())
	}

	p.state = stateActive
	p.started = time.Now()
	p.agent.Logger().Printf("Started module [ %s ]", p.Name())

	return nil
}

func (p *ClientPlugin) loadClientSymbols() error {

	raw, err := plugin.Open(p.filename)
	if err != nil {
		return fmt.Errorf("plugin load error: Failed to open file [ %s ] to load symbols - %s", p.filename, err)
	}

	if err := p.loadClientSymbol(raw, "Init"); err != nil {
		return err
	}

	// Add more symbols here...

	return nil
}

func (p *ClientPlugin) loadClientSymbol(rawPlugin *plugin.Plugin, SymbolName string) error {

	s, err := rawPlugin.Lookup(SymbolName)
	if err != nil {
		return fmt.Errorf("plugin load error: Failed to find Symbol [ %s ] in file [ %s ] - %s", SymbolName, p.filename, err)
	}

	// Dispatch on the type of the returned symbol.
	// Each of the expected or possible symbols must have a corresponding case
	// As more possible symbols are added, simply extend the logic to account
	// for all the types.
	switch s.(type) {
	case ClientInitFunc:
		p.init = s.(ClientInitFunc)
	default:
		return fmt.Errorf("plugin load error: Unknown type returned for symbol [ %s ], got [ %s ]", SymbolName, getFuncSignature(s))
	}

	return nil
}

func (p *ClientPlugin) unloadSymbols() {

	p.mu.Lock()
	defer p.mu.Unlock()

	p.state = stateNotLoaded

	p.unloadDefaultSymbols()

	p.init = nil
	// ...
}

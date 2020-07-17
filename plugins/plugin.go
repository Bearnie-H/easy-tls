package plugins

import (
	"errors"
	"fmt"
	"log"
	"plugin"
	"reflect"
	"strings"
	"sync"
	"time"
)

// Module is the interface which must be satisfied by any plugins intended
// to be managed by this package. This is distinct from the API Contract
// the module must satisfy, which refers to the set of functions and signatures
// the module must export. This interface refers to how the Plugin Agent will
// communicate with a module, regardless of what these actual exported
// signatures are.
type Module interface {

	// These functions will be accessible via the Command-Server interface

	Start() error
	GetVersion() (*SemanticVersion, error)
	Uptime() time.Duration
	Reload() error
	Stop() error
	State() PluginState

	// These functions will only be accessible and used internally.

	Load() error
	Name() string
	ReadStatus() error
	Done() <-chan struct{}
	AddArguments(...interface{})
}

type (
	// StatusFunc is the defined type of the Status function exported by all compliant plugins.
	StatusFunc = func() (<-chan PluginStatus, error)

	// VersionFunc is the defined type of the Version function exported by all compliant plugins.
	VersionFunc = func(SemanticVersion) error

	// NameFunc is the defined type of the Name function exported by all compliant plugins.
	NameFunc = func() string

	// StopFunc is the defined type of the Stop function exported by all compliant plugins.
	StopFunc = func() error

	// InitFunc is the defined type of the most generic Init function exported by generic plugins.
	// This does not need to be exported, as long as SOME Init() function is exported, and whatever
	// implementation of the Module interface is used, properly extracts this symbol.
	InitFunc = func(...interface{}) error
)

// PluginState represents the current state of a plugin
type PluginState int

const (
	stateNotLoaded PluginState = iota
	stateLoaded
	stateActive
)

func (S PluginState) String() string {
	switch S {
	case stateNotLoaded:
		return "Not Loaded"
	case stateLoaded:
		return "Stopped"
	case stateActive:
		return "Running"
	default:
		return "Unknown"
	}
}

// GenericPlugin implements a basic version of the Module interface, specifically the
// set of parameters and methods which are generic and common across either
// client, server, or other modules.
type GenericPlugin struct {

	// The name of the file to attempt to load as a module
	filename string

	// Reference to the agent this plugin is loaded into
	agent *Agent

	// The additional arguments to pass to Init().
	args []interface{}

	// The channel to fire on when the module finished executing.
	done chan struct{}

	// The logger to write messages to from the module.
	logger *log.Logger

	// Protects state
	mu    *sync.Mutex
	state PluginState

	// When was the module most recently started
	started time.Time

	// Common functions exported by all plugins
	status  StatusFunc
	version VersionFunc
	name    NameFunc
	stop    StopFunc

	// Generic Init() function, if no other more customized version is given.
	init InitFunc
}

// GetVersion will return the version information of the plugin, as reported
// in the canonical name,
func (p *GenericPlugin) GetVersion() (*SemanticVersion, error) {

	x := strings.Split(p.Name(), "-")
	rawVersion := x[len(x)-1]
	rawVersion = strings.Split(rawVersion, " ")[0]

	return ParseVersion(rawVersion)
}

// Name satisfies the Name() component of the Module interface by calling the generic
// Name() function exported by the plugins
func (p *GenericPlugin) Name() string {

	switch p.state {
	case stateNotLoaded:
		return "<Module Not Loaded>"
	default:
	}

	return p.name()
}

// Load implements the most generic Load functionality for a generic plugin.
// This will typically be overloaded/superceded by the more specific implementation
// of the Module interface.
func (p *GenericPlugin) Load() error {

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

	p.state = stateLoaded

	p.agent.Logger().Printf("Loaded symbols for module [ %s ]", p.Name())

	return nil
}

// Start implements the most generic Start functionality for a generic plugin.
// This will typically be overloaded/superceded by the more specific implementation
// of the Module interface.
func (p *GenericPlugin) Start() error {

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

	p.state = stateActive
	p.started = time.Now()

	switch {
	case p.init != nil:
		{
			if err := p.init(p.args...); err != nil {
				p.stop()
				p.state = stateLoaded
				return err
			}
		}
	default:
		return fmt.Errorf("plugin error: No Init() function loaded for module [ %s ]", p.Name())
	}

	p.agent.Logger().Printf("Started module [ %s ]", p.Name())

	return nil
}

// Reload implements the most generic Reload functionality for a generic plugin.
// This will typically be overloaded/superceded by the more specific implementation
// of the Module interface.
func (p *GenericPlugin) Reload() error {

	var err error

	if err = p.Stop(); err != nil {
		p.agent.Logger().Printf("plugin reload error: Error stopping plugin [ %s ] for reload - %s", p.Name(), err)
		return err
	}

	p.unloadDefaultSymbols()

	if err = p.Load(); err != nil {
		p.agent.Logger().Printf("plugin reload error: Error loading plugin [ %s ] for reload - %s", p.Name(), err)
		return err
	}

	if err = p.Start(); err != nil {
		p.agent.Logger().Printf("plugin reload error: Error starting plugin [ %s ] for reload - %s", p.Name(), err)
		return err
	}

	return nil
}

// AddArguments will append the given arguments to the set of arguments to pass in to the plugin
func (p *GenericPlugin) AddArguments(args ...interface{}) {
	p.args = append(p.args, args...)
}

// ReadStatus wiill spawn a go-routine to read messages off the Status() channel
// from the plugin until it is closed.
func (p *GenericPlugin) ReadStatus() error {

	switch p.state {
	case stateNotLoaded:
		return errors.New("plugin error: Plugin symbols not loaded")
	default:
	}

	C, err := p.status()
	if err != nil {
		return err
	}
	if C == nil {
		return errors.New("plugin error: Module returned nil status channel")
	}

	p.done = make(chan struct{})

	go func(p *GenericPlugin, C <-chan PluginStatus) {

		// This go-routine can only ever exit if the plugin closes the send half of the StatusChannel
		// This only happens at the end of a Stop(), which allows closing of this channel to
		// indicate the plugin is stopped.
		defer func() {
			p.agent.Logger().Printf("Finished logging for module [ %s ]", p.Name())
			p.done <- struct{}{}
			p.mu.Lock()
			p.state = stateLoaded
			p.mu.Unlock()
			close(p.done)
		}()

		// Range the status channel until it's closed by the module
		for M := range C {

			// If the message is a fatal error, kill the plugin
			if M.fatal {
				// This needs to use the exported function, not the GenericPlugin
				// Stop() method to avoid a deadlock condition on the p.done
				// channel
				p.stop()
			}
			p.logger.Println(M.String())
		}

	}(p, C)

	p.agent.Logger().Printf("Started logging for module [ %s ]", p.Name())

	return nil
}

// Uptime will return how long a module has been active for
func (p *GenericPlugin) Uptime() time.Duration {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.state == stateActive {
		return time.Now().Sub(p.started)
	}
	return 0
}

// Stop implements the Stop method of the Module interface.
// This will stop the logic of the plugin according to the exported Stop()
// function, and will wait for ALL status messages to be written out before
// returning any errors which occured during shutdown
func (p *GenericPlugin) Stop() error {

	p.agent.Logger().Printf("Stopping module [ %s ]", p.Name())

	switch p.state {
	case stateNotLoaded:
		p.agent.Logger().Printf("Module [ %s ] not loaded, nothing to stop", p.Name())
		return nil
	case stateLoaded:
		p.agent.Logger().Printf("Module [ %s ] already stopped", p.Name())
		return nil
	default:
	}

	defer func(p *GenericPlugin) {
		<-p.Done()
		p.mu.Lock()
		p.state = stateLoaded
		p.mu.Unlock()
		p.agent.Logger().Printf("Stopped module [ %s ]", p.Name())
	}(p)

	return p.stop()
}

// State exposes the internal state variable, so it can be queried.
func (p *GenericPlugin) State() PluginState {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.state
}

// Done will return a channel that fires once then closes to indicate
// that a plugin is finished.
func (p *GenericPlugin) Done() <-chan struct{} {
	return p.done
}

// LoadDefaultSymbols will load the generic, default symbols for the following
// functions into the plugin, returning any errors along the way:
//	Status
//	Version
//	Name
//	Stop
//	Init (Not necessarily exported with the default signature)
func (p *GenericPlugin) loadDefaultSymbols() error {

	raw, err := plugin.Open(p.filename)
	if err != nil {
		return fmt.Errorf("plugin load error: Failed to open file [ %s ] to load symbols - %s", p.filename, err)
	}

	if err := p.loadDefaultSymbol(raw, "Status"); err != nil {
		return err
	}

	if err := p.loadDefaultSymbol(raw, "Version"); err != nil {
		return err
	}

	if err := p.loadDefaultSymbol(raw, "Name"); err != nil {
		return err
	}

	if err := p.loadDefaultSymbol(raw, "Stop"); err != nil {
		return err
	}

	// Attempt to load a default Init() symbol, allowing this to fail if a more customized Init() symbol is present.
	p.loadDefaultSymbol(raw, "Init")

	// ...

	return nil
}

func (p *GenericPlugin) loadDefaultSymbol(rawPlugin *plugin.Plugin, SymbolName string) error {

	// Look for the exported symbol name
	s, err := rawPlugin.Lookup(SymbolName)
	if err != nil {
		return fmt.Errorf("plugin load error: Failed to find Symbol [ %s ] in file [ %s ] - %s", SymbolName, p.filename, err)
	}

	// Typeswitch to set the corresponding symbol in the plugin
	switch s.(type) {
	case StatusFunc:
		p.status = s.(StatusFunc)
	case VersionFunc:
		p.version = s.(VersionFunc)
	case NameFunc:
		p.name = s.(NameFunc)
	case StopFunc:
		p.stop = s.(StopFunc)
	case InitFunc:
		p.init = s.(InitFunc)
	default:
		return fmt.Errorf("plugin load error: Unknown type returned for symbol [ %s ] in file [ %s ], got [ %s ]", SymbolName, p.filename, getFuncSignature(s))
	}

	return nil
}

func (p *GenericPlugin) unloadDefaultSymbols() {
	p.name = nil
	p.status = nil
	p.stop = nil
	p.version = nil
	p.init = nil
}

func getFuncSignature(f interface{}) string {
	t := reflect.TypeOf(f)
	if t.Kind() != reflect.Func {
		return "<not a function>"
	}

	buf := strings.Builder{}
	buf.WriteString("func (")
	for i := 0; i < t.NumIn(); i++ {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(t.In(i).String())
	}
	buf.WriteString(")")
	if numOut := t.NumOut(); numOut > 0 {
		if numOut > 1 {
			buf.WriteString(" (")
		} else {
			buf.WriteString(" ")
		}
		for i := 0; i < t.NumOut(); i++ {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(t.Out(i).String())
		}
		if numOut > 1 {
			buf.WriteString(")")
		}
	}

	return buf.String()
}

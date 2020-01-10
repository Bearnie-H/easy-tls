package plugins

import (
	"errors"
	"fmt"
	"plugin"
	"time"
)

// Plugin represents the most generic features and functionality of a Plugin Object
type Plugin struct {

	// Filename represents purely the filename component of the plugin file.
	Filename string

	// Filepath represents the full path to the plugin file.
	Filepath string

	// Plugins are allowed to have input arguments, which can be stored here to be passed in by the pluginAgent
	InputArguments []interface{}

	// All Plugins implement the basic API Contract.
	// This must be a struct and not an interface because the actual function bodies will be returned from loading the plugin file.
	PluginAPI
}

// SetInputArguments will set the input arguments passed to a given plugin to be exactly what is passed to this function.
func (P *Plugin) SetInputArguments(args ...interface{}) {
	P.InputArguments = args
}

// PluginAPI represents the base API contract which must be satisfied by ANY plugin.
type PluginAPI struct {

	// Query the current status of the plugin.
	//
	// This must return an output-only unbuffered channel, allowing the plugin to directly send status messages as they are generated.
	// If this channel is not read from, it will not block itself, and will only present the most recent message.
	Status func() (<-chan PluginStatus, error)

	// Query the Semantic Versioning compatabilities of the plugin.
	//
	// This will accept the Semantic Version of the Plugin at hand and compare it against it's set of acceptable framework versions.  A nil error implies compatability.
	Version func(SemanticVersion) error

	// Query the Name of the Plugin.
	//
	// This must return the name of the plugin, in canonical format.
	Name func() string

	// Stop the plugin.
	//
	// This must trigger a full stop of any internal behaviours of the plugin, only returning once ALL internal behaviours have halted.  This should return any and all errors which arise during shutdown and are not explicitly handled by the shutdown.  The Agent makes no guarantee on how long AFTER receiving the return value from this call the application will run for, so this must represent the FINAL valid state of a plugin.
	Stop func() error
}

// PluginStatus represents a single status message from a given EasyTLS-compliant plugin.
type PluginStatus struct {
	Message string
	Error   error
	IsFatal bool
}

func (S PluginStatus) String() string {

	if S.IsFatal {
		return time.Now().Format("2006/01/02 15:04:05") + " " + fmt.Sprintf("FATAL ERROR: %s - %s\n", S.Message, S.Error)
	}

	if S.Error != nil {
		return time.Now().Format("2006/01/02 15:04:05") + " " + fmt.Sprintf("Warning: %s - %s\n", S.Message, S.Error)
	}

	return time.Now().Format("2006/01/02 15:04:05") + " " + S.Message + "\n"
}

func loadDefaultPluginSymbols(Filename string) (PluginAPI, error) {
	API := PluginAPI{}

	rawPlug, err := plugin.Open(Filename)
	if err != nil {
		return API, err
	}

	if API.Status, err = loadStatusSymbol(rawPlug); err != nil {
		return API, err
	}

	if API.Version, err = loadVersionSymbol(rawPlug); err != nil {
		return API, err
	}

	if API.Name, err = loadNameSymbol(rawPlug); err != nil {
		return API, err
	}

	if API.Stop, err = loadStopSymbol(rawPlug); err != nil {
		return API, err
	}

	return API, nil
}

func loadStatusSymbol(p *plugin.Plugin) (func() (<-chan PluginStatus, error), error) {
	sym, err := p.Lookup("Status")
	if err != nil {
		return nil, err
	}

	StatusSymbol, ok := sym.(func() (<-chan PluginStatus, error))
	if !ok {
		return nil, errors.New("easytls plugin error: Status symbol has invalid signature")
	}

	return StatusSymbol, nil
}

func loadVersionSymbol(p *plugin.Plugin) (func(SemanticVersion) error, error) {
	sym, err := p.Lookup("Version")
	if err != nil {
		return nil, err
	}

	VersionSymbol, ok := sym.(func(SemanticVersion) error)
	if !ok {
		return nil, errors.New("easytls plugin error: Version symbol has invalid signature")
	}

	return VersionSymbol, nil
}

func loadNameSymbol(p *plugin.Plugin) (func() string, error) {
	sym, err := p.Lookup("Name")
	if err != nil {
		return nil, err
	}

	NameSymbol, ok := sym.(func() string)
	if !ok {
		return nil, errors.New("easytls plugin error: Name symbol has invalid signature")
	}

	return NameSymbol, nil
}

func loadStopSymbol(p *plugin.Plugin) (func() error, error) {
	sym, err := p.Lookup("Stop")
	if err != nil {
		return nil, err
	}

	StopSymbol, ok := sym.(func() error)
	if !ok {
		return nil, errors.New("easytls plugin error: Stop symbol has invalid signature")
	}

	return StopSymbol, nil
}

package plugins

import (
	"errors"
	"plugin"
)

// PluginAPI represents the base API contract which must be satisfied by ANY plugin.
type PluginAPI struct {

	// Query the current status of the plugin.
	//
	// This must return an output-only channel, allowing the plugin to directly send status messages as they are generated.
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

func loadStatusSymbol(p *plugin.Plugin) (StatusFunc, error) {
	sym, err := p.Lookup("Status")
	if err != nil {
		return nil, err
	}

	StatusSymbol, ok := sym.(StatusFunc)
	if !ok {
		return nil, errors.New("easytls plugin error: Status symbol has invalid signature")
	}

	return StatusSymbol, nil
}

func loadVersionSymbol(p *plugin.Plugin) (VersionFunc, error) {
	sym, err := p.Lookup("Version")
	if err != nil {
		return nil, err
	}

	VersionSymbol, ok := sym.(VersionFunc)
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

	NameSymbol, ok := sym.(NameFunc)
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

	StopSymbol, ok := sym.(StopFunc)
	if !ok {
		return nil, errors.New("easytls plugin error: Stop symbol has invalid signature")
	}

	return StopSymbol, nil
}

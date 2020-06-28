package plugins

import (
	"fmt"
	"plugin"
	"reflect"
	"strings"
)

// PluginAPI represents the base API contract which must be satisfied by ANY plugin.
type PluginAPI struct {

	// Query the current status of the plugin.
	//
	// This must return an output-only channel, allowing the plugin to directly send status messages as they are generated.
	Status StatusFunc

	// Query the Semantic Versioning compatabilities of the plugin.
	//
	// This will accept the Semantic Version of the Plugin at hand and compare it against it's set of acceptable framework versions.  A nil error implies compatability.
	Version VersionFunc

	// Query the Name of the Plugin.
	//
	// This must return the name of the plugin, in canonical format.
	Name NameFunc

	// Stop the plugin.
	//
	// This must trigger a full stop of any internal behaviours of the plugin, only returning once ALL internal behaviours have halted.  This should return any and all errors which arise during shutdown and are not explicitly handled by the shutdown.  The Agent makes no guarantee on how long AFTER receiving the return value from this call the application will run for, so this must represent the FINAL valid state of a plugin.
	Stop StopFunc
}

func (P *Plugin) loadDefaultPluginSymbols(Filename string) error {

	rawPlug, err := plugin.Open(Filename)
	if err != nil {
		return err
	}

	if err := P.PluginAPI.loadStatusSymbol(rawPlug); err != nil {
		return err
	}

	if err := P.PluginAPI.loadVersionSymbol(rawPlug); err != nil {
		return err
	}

	if err := P.PluginAPI.loadNameSymbol(rawPlug); err != nil {
		return err
	}

	if err := P.PluginAPI.loadStopSymbol(rawPlug); err != nil {
		return err
	}

	return nil
}

func (A *PluginAPI) loadStatusSymbol(p *plugin.Plugin) error {

	SymbolName := "Status"
	sym, err := p.Lookup(SymbolName)
	if err != nil {
		return err
	}

	var ok bool
	A.Status, ok = sym.(StatusFunc)
	if !ok {
		return fmt.Errorf("easytls plugin error: Invalid %s() signature, expected [ %s ] - got [ %s ]", SymbolName, getFuncSignature(A.Status), getFuncSignature(sym))
	}

	return nil
}

func (A *PluginAPI) loadVersionSymbol(p *plugin.Plugin) error {

	SymbolName := "Version"
	sym, err := p.Lookup(SymbolName)
	if err != nil {
		return err
	}

	var ok bool
	A.Version, ok = sym.(VersionFunc)
	if !ok {
		return fmt.Errorf("easytls plugin error: Invalid %s() signature, expected [ %s ] - got [ %s ]", SymbolName, getFuncSignature(A.Version), getFuncSignature(sym))
	}

	return nil
}

func (A *PluginAPI) loadNameSymbol(p *plugin.Plugin) error {

	SymbolName := "Name"
	sym, err := p.Lookup(SymbolName)
	if err != nil {
		return err
	}

	var ok bool
	A.Name, ok = sym.(NameFunc)
	if !ok {
		return fmt.Errorf("easytls plugin error: Invalid %s() signature, expected [ %s ] - got [ %s ]", SymbolName, getFuncSignature(A.Name), getFuncSignature(sym))
	}

	return nil
}

func (A *PluginAPI) loadStopSymbol(p *plugin.Plugin) error {

	SymbolName := "Stop"
	sym, err := p.Lookup(SymbolName)
	if err != nil {
		return err
	}

	var ok bool
	A.Stop, ok = sym.(StopFunc)
	if !ok {
		return fmt.Errorf("easytls plugin error: Invalid %s() signature, expected [ %s ] - got [ %s ]", SymbolName, getFuncSignature(A.Stop), getFuncSignature(sym))
	}

	return nil
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

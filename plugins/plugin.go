package plugins

import (
	"log"
	"path"
)

type (
	// StatusFunc is the defined type of the Status function exported by all compliant plugins.
	StatusFunc = func() (<-chan PluginStatus, error)

	// VersionFunc is the defined type of the Version function exported by all compliant plugins.
	VersionFunc = func(SemanticVersion) error

	// NameFunc is the defined type of the Name function exported by all compliant plugins.
	NameFunc = func() string

	// StopFunc is the defined type of the Stop function exported by all compliant plugins.
	StopFunc = func() error
)

// Plugin represents the most generic features and functionality of a Plugin Object
type Plugin struct {

	// Filename represents purely the filename component of the plugin file.
	Filename string

	// Filepath represents the full path to the plugin file.
	Filepath string

	// Plugins are allowed to have input arguments, which can be stored here to be passed in by the pluginAgent
	InputArguments []interface{}

	// Done will fire only when a plugin is fully stopped and all status messages have been read.
	Done chan struct{}

	// The logger the plugin should use.
	Logger *log.Logger

	// All Plugins implement the basic API Contract.
	// This must be a struct and not an interface because the actual function bodies will be returned from loading the plugin file.
	PluginAPI
}

// NewPlugin is the canonical method to create a new Plugin object.
func NewPlugin(Filepath string, Logger *log.Logger, DefaultArguments ...interface{}) Plugin {
	return Plugin{
		Filename:       path.Base(Filepath),
		Filepath:       path.Dir(Filepath),
		InputArguments: DefaultArguments,
		Done:           make(chan struct{}),
		Logger:         Logger,
	}
}

// SetInputArguments will set the input arguments passed to a given plugin to be exactly what is passed to this function.
func (P *Plugin) SetInputArguments(args ...interface{}) {
	P.InputArguments = args
}

// AppendInputArguments will append a set of arguments to the list which will be provided to the plugin when the Init symbol is called.
func (P *Plugin) AppendInputArguments(args ...interface{}) {
	P.InputArguments = append(P.InputArguments, args...)
}

// Kill will kill a generic plugin, stopping the internal logic and waiting for all of the status messages to be read out.
func (P *Plugin) Kill() error {
	err := P.Stop()
	<-P.Done
	return err
}

func (P *Plugin) readStatus() error {
	StatusChannel, err := P.Status()
	if err != nil {
		return err
	}

	go func(P *Plugin, StatusChannel <-chan PluginStatus) {

		// This go-routine can only ever exit if the plugin closes the send half of the StatusChannel
		// This only happens at the end of a Stop(), which allows closing of this channel to
		// indicate the plugin is stopped.
		defer func() { P.Done <- struct{}{}; close(P.Done) }()

		// Read the status channel until it's closed by the plugin.
		for Message := range StatusChannel {

			// If the message is fatal, kill the plugin
			if Message.fatal {
				go P.Kill()
			}
			P.Logger.Println(Message.String())
		}
	}(P, StatusChannel)

	return nil
}

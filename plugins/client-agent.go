package plugins

import (
	"errors"
	"fmt"
	"log"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Bearnie-H/easy-tls/client"
)

// ClientPluginAgent represents the a Plugin Manager agent to be used with a SimpleClient.
type ClientPluginAgent struct {
	RegisteredPlugins map[string]*ClientPlugin

	client             *client.SimpleClient
	frameworkVersion   SemanticVersion
	logger             *log.Logger
	PluginSearchFolder string
}

// NewClientAgent will create a new Client Plugin agent, ready to register plugins.
func NewClientAgent(Client *client.SimpleClient, PluginFolder string) (*ClientPluginAgent, error) {

	var err error
	if Client == nil {
		Client, err = client.NewClientHTTP()
		if err != nil {
			return nil, err
		}
	}

	A := &ClientPluginAgent{
		client:             Client,
		frameworkVersion:   ClientFrameworkVersion,
		PluginSearchFolder: PluginFolder,
		RegisteredPlugins:  make(map[string]*ClientPlugin),
		logger:             Client.Logger(),
	}

	return A, nil
}

// GetPluginByName will return a pointer to the requested plugin.  This is typically used to provide input arguments for when the plugin is Initiated.
func (CA *ClientPluginAgent) GetPluginByName(Name string) (*ClientPlugin, error) {

	Name = strings.ToLower(Name)
	for index, p := range CA.RegisteredPlugins {
		name := strings.ToLower(p.Name())
		if strings.HasPrefix(name, Name) {
			return CA.RegisteredPlugins[index], nil
		}
	}
	return nil, fmt.Errorf("easytls plugin error: Failed to find plugin %s", Name)
}

// StopPluginByName will attempt to stop a given plugin by name, if it exists
func (CA *ClientPluginAgent) StopPluginByName(Name string) error {

	p, err := CA.GetPluginByName(Name)
	if err != nil {
		return err
	}
	return p.Kill()
}

// RegisterPlugins will configure and register all of the plugins in the previously specified PluginFolder.  This will not start any of the plugins, but will only load the necessary symbols from them.
func (CA *ClientPluginAgent) RegisterPlugins() error {

	// Search for all plugins in the designated search folder...
	files, err := filepath.Glob(path.Join(CA.PluginSearchFolder, "*.so"))
	if err != nil {
		return err
	}

	// Sort all files alphabetically, to assert a standard import order
	sort.Strings(files)

	var loadErrors error

	// Attempt to load all of the plugins.
	for index := range files {
		newPlugin, err := InitializeClientPlugin(files[index], CA.frameworkVersion, CA.client)
		if err == nil {
			CA.RegisteredPlugins[newPlugin.Name()] = newPlugin
		} else {
			if loadErrors == nil {
				loadErrors = fmt.Errorf("easytls plugin agent error: %s", err)
			} else {
				loadErrors = fmt.Errorf("%s\n%s", loadErrors, err)
			}
		}
	}

	return loadErrors
}

// Run will start the ClientPlugin Agent, starting each of the registered plugins.
// blocking represents if the rest of the application should block on this call or not.
func (CA *ClientPluginAgent) Run(blocking bool) error {

	if blocking {
		return CA.run()
	}

	go CA.run()
	return nil
}

func (CA *ClientPluginAgent) run() error {

	if len(CA.RegisteredPlugins) == 0 {
		return errors.New("easytls plugin error: Client Plugin Agent has 0 registered plugins")
	}

	for pluginName, registeredPlugin := range CA.RegisteredPlugins {
		if err := registeredPlugin.readStatus(); err != nil {
			CA.logger.Printf("easytls plugin error: Failed to start status logging for plugin [ %s ] - %s", pluginName, err)
			continue
		}

		// Start the plugin...
		registeredPlugin.Start()
	}

	CA.Wait()

	return nil
}

// Stop will cause ALL of the currently Running Plugins to safely stop.
func (CA *ClientPluginAgent) Stop() error {

	FailedPlugins := []string{}

	for _, P := range CA.RegisteredPlugins {
		if err := P.Kill(); err != nil {
			FailedPlugins = append(FailedPlugins, P.Name())
		}
	}

	if len(FailedPlugins) > 0 {
		return fmt.Errorf("easytls plugin agent error: Error(s) occured while shutting down plugins %v", FailedPlugins)
	}

	return nil
}

// Wait for the plugin agent to stop safely.
func (CA *ClientPluginAgent) Wait() {
	for _, P := range CA.RegisteredPlugins {
		<-P.Done
	}
}

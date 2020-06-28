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
func (A *ClientPluginAgent) GetPluginByName(Name string) (*ClientPlugin, error) {

	Name = strings.ToLower(Name)
	for index, p := range A.RegisteredPlugins {
		name := strings.ToLower(p.Name())
		if strings.HasPrefix(name, Name) {
			return A.RegisteredPlugins[index], nil
		}
	}
	return nil, fmt.Errorf("easytls plugin error: Failed to find plugin %s", Name)
}

// StopPluginByName will attempt to stop a given plugin by name, if it exists
func (A *ClientPluginAgent) StopPluginByName(Name string) error {

	p, err := A.GetPluginByName(Name)
	if err != nil {
		return err
	}
	return p.Kill()
}

// AddPluginArguments will configure the registered plugins to all receive a copy of args
// when their Init() function is called.
func (A *ClientPluginAgent) AddPluginArguments(args ...interface{}) {
	for _, P := range A.RegisteredPlugins {
		P.AppendInputArguments(args...)
	}
}

// RegisterPlugins will configure and register all of the plugins in the previously specified PluginFolder.  This will not start any of the plugins, but will only load the necessary symbols from them.
func (A *ClientPluginAgent) RegisterPlugins() error {

	// Search for all plugins in the designated search folder...
	files, err := filepath.Glob(path.Join(A.PluginSearchFolder, "*.so"))
	if err != nil {
		return err
	}

	// Sort all files alphabetically, to assert a standard import order
	sort.Strings(files)

	var loadErrors error

	// Attempt to load all of the plugins.
	for index := range files {
		newPlugin, err := InitializeClientPlugin(files[index], A.frameworkVersion, A.client)
		if err == nil {
			A.RegisteredPlugins[newPlugin.Name()] = newPlugin
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
func (A *ClientPluginAgent) Run(blocking bool) error {

	if blocking {
		return A.run()
	}

	go A.run()
	return nil
}

func (A *ClientPluginAgent) run() error {

	if len(A.RegisteredPlugins) == 0 {
		return errors.New("easytls plugin error: Client Plugin Agent has 0 registered plugins")
	}

	for pluginName, registeredPlugin := range A.RegisteredPlugins {
		if err := registeredPlugin.readStatus(); err != nil {
			A.logger.Printf("easytls plugin error: Failed to start status logging for plugin [ %s ] - %s", pluginName, err)
			continue
		}

		// Start the plugin...
		registeredPlugin.Start()
	}

	A.Wait()

	return nil
}

// Stop will cause ALL of the currently Running Plugins to safely stop.
func (A *ClientPluginAgent) Stop() error {

	FailedPlugins := []string{}

	for _, P := range A.RegisteredPlugins {
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
func (A *ClientPluginAgent) Wait() {
	for _, P := range A.RegisteredPlugins {
		<-P.Done
	}
}

func (A *ClientPluginAgent) sortedPluginNames() []string {
	Names := []string{}
	for n := range A.RegisteredPlugins {
		Names = append(Names, n)
	}
	sort.Strings(Names)
	return Names
}

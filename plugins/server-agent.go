package plugins

import (
	"errors"
	"fmt"
	"log"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Bearnie-H/easy-tls/server"
	"github.com/gorilla/mux"
)

// ServerPluginAgent represents the a Plugin Manager agent to be used with a SimpleServer.
type ServerPluginAgent struct {
	frameworkVersion   SemanticVersion
	URLBase            string
	RegisteredPlugins  map[string]*ServerPlugin
	logger             *log.Logger
	PluginSearchFolder string
	server             *server.SimpleServer
	router             *mux.Router
}

// NewServerAgent will create a new Server Plugin agent, ready to register plugins.
// A nil logger defaults to the standard logger provided by the log package.
//
// The URLBase allows for the plugins to base the handlers they export at a
// specific Base URL.
func NewServerAgent(PluginFolder, URLBase string, Server *server.SimpleServer) (*ServerPluginAgent, error) {

	if Server == nil {
		var err error
		Server, err = server.NewServerHTTP()
		if err != nil {
			return nil, err
		}
	}

	var Router *mux.Router

	if !strings.HasSuffix(URLBase, "/") {
		URLBase += "/"
	}

	// If the URLBase is serving from the root, just use the primary server router
	// as a special case; otherwise create a new sub-router for the agent explicitly.
	if URLBase == "/" {
		Router = Server.Router()
	} else {
		Router = Server.Router().PathPrefix(URLBase).Subrouter()
	}

	A := &ServerPluginAgent{
		frameworkVersion:   ServerFrameworkVersion,
		URLBase:            URLBase,
		PluginSearchFolder: PluginFolder,
		RegisteredPlugins:  make(map[string]*ServerPlugin),
		logger:             Server.Logger(),
		server:             Server,
		router:             Router,
	}

	return A, nil
}

// GetPluginByName will return a pointer to the requested plugin.  This is typically used to provide input arguments for when the plugin is Initiated.
func (A *ServerPluginAgent) GetPluginByName(Name string) (*ServerPlugin, error) {
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
func (A *ServerPluginAgent) StopPluginByName(Name string) error {

	p, err := A.GetPluginByName(Name)
	if err != nil {
		return err
	}
	return p.Kill()
}

// RegisterPlugins will configure and register all of the plugins in the previously specified PluginFolder.  This will not start any of the plugins, but will only load the necessary symbols from them.
func (A *ServerPluginAgent) RegisterPlugins() error {

	// Search for all plugins in the designated search folder...
	files, err := filepath.Glob(path.Join(A.PluginSearchFolder, "*.so"))
	if err != nil {
		return err
	}

	// Sort all files alphabetically, to assert a standard import order
	sort.Strings(files)

	var loadErrors error

	// Attempt to load all of the plugins.
	for _, f := range files {
		newPlugin, err := InitializeServerPlugin(f, A.frameworkVersion, A.logger)
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

// LoadRoutes will load routes from all of the registered plugins into the Server.
func (A *ServerPluginAgent) LoadRoutes() error {

	// Walk the set of registered plugins, adding the routes from each to the router.
	// Do this in name-sorted order, to assert a definite registration order, rather
	// than allowing for arbitrary order by ranging the map directly
	for _, Name := range A.sortedPluginNames() {
		p := A.RegisteredPlugins[Name]
		// Call the correct Init() function, based on what was loaded during registration
		switch {
		case p.InitHandlers != nil:
			routes, err := p.InitHandlers(p.InputArguments...)
			if err != nil {
				return err
			}
			for i := range routes {
				routes[i].Path = A.URLBase + routes[i].Path
			}
			A.server.AddHandlers(A.router, routes...)
		case p.InitSubrouter != nil:
			routes, prefix, err := p.InitSubrouter(p.InputArguments...)
			if err != nil {
				return err
			}
			A.server.AddSubrouter(A.router, A.URLBase+prefix, routes...)
		default:
			return errors.New("easytls plugin error: Plugin exported no Init() function, no routes loaded")
		}
	}

	return nil
}

// AddPluginArguments will configure the registered plugins to all receive a copy of args
// when their Init() function is called.
func (A *ServerPluginAgent) AddPluginArguments(args ...interface{}) {
	for _, P := range A.RegisteredPlugins {
		P.AppendInputArguments(args...)
	}
}

// Run will start the ServerPlugin Agent, starting each of the registered plugins.
// blocking represents if the rest of the application should block on this SAll or not.
func (A *ServerPluginAgent) Run(blocking bool) error {

	if blocking {
		return A.run()
	}

	go A.run()
	return nil
}

func (A *ServerPluginAgent) run() error {

	if len(A.RegisteredPlugins) == 0 {
		return errors.New("easytls plugin agent error: Server Plugin Agent has 0 registered plugins")
	}

	for pluginName, registeredPlugin := range A.RegisteredPlugins {
		if err := registeredPlugin.readStatus(); err != nil {
			A.logger.Printf("easytls plugin error: Failed to start status logging for plugin [ %s ] - %s", pluginName, err)
			continue
		}
	}

	A.Wait()

	return nil
}

// Stop will cause ALL of the currently Running Plugins to safely stop.
func (A *ServerPluginAgent) Stop() error {

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
func (A *ServerPluginAgent) Wait() {
	for _, P := range A.RegisteredPlugins {
		<-P.Done
	}
}

func (A *ServerPluginAgent) sortedPluginNames() []string {
	Names := []string{}
	for n := range A.RegisteredPlugins {
		Names = append(Names, n)
	}
	sort.Strings(Names)
	return Names
}

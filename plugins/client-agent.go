package plugins

import (
	"errors"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Bearnie-H/easy-tls/client"
)

// ClientPluginAgent represents the a Plugin Manager agent to be used with a SimpleClient.
type ClientPluginAgent struct {
	client             *client.SimpleClient
	frameworkVersion   SemanticVersion
	RegisteredPlugins  []ClientPlugin
	logger             io.WriteCloser
	PluginSearchFolder string

	mux     sync.Mutex
	stopped atomic.Value
}

// NewClientAgent will create a new Client Plugin agent, ready to register plugins.
func NewClientAgent(Client *client.SimpleClient, PluginFolder string, Logger io.WriteCloser) (*ClientPluginAgent, error) {

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
		RegisteredPlugins:  []ClientPlugin{},
		logger:             Logger,
		stopped:            atomic.Value{},
		mux:                sync.Mutex{},
	}
	A.stopped.Store(false)

	return A, nil
}

// GetPluginByName will return a pointer to the requested plugin.  This is typically used to provide input arguments for when the plugin is Initiated.
func (CA *ClientPluginAgent) GetPluginByName(Name string) (*ClientPlugin, error) {
	Name = strings.ToLower(Name)
	for index, p := range CA.RegisteredPlugins {
		name := strings.ToLower(p.Name())
		if strings.HasPrefix(name, Name) {
			return &(CA.RegisteredPlugins[index]), nil
		}
	}
	return nil, fmt.Errorf("easytls plugin error - Failed to find plugin %s", Name)
}

// StopPluginByName will attempt to stop a given plugin by name, if it exists
func (CA *ClientPluginAgent) StopPluginByName(Name string) error {

	p, err := CA.GetPluginByName(Name)
	if err != nil {
		return err
	}
	return p.Stop()
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
	for _, f := range files {
		newPlugin, err := InitializeClientPlugin(f, CA.frameworkVersion)
		if err == nil {
			CA.RegisteredPlugins = append(CA.RegisteredPlugins, *newPlugin)
		} else {
			if loadErrors == nil {
				loadErrors = fmt.Errorf("easytls plugin agent error - %s", err)
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
		return errors.New("easytls plugin error - Client Plugin Agent has 0 registered plugins")
	}

	wg := &sync.WaitGroup{}
	for _, registeredPlugin := range CA.RegisteredPlugins {
		wg.Add(1)

		// Start the plugin...
		go func(c *client.SimpleClient, p ClientPlugin, wg *sync.WaitGroup) {

			// Extract the status channel
			statusChan, err := p.Status()

			// An error retrieving the status channel stops the logging.
			if err != nil {
				CA.logger.Write([]byte(err.Error() + "\n"))
				wg.Done()
				return
			}

			// Start logging plugin status messages.
			go func(wg *sync.WaitGroup) {

				// If the plugin exits, decrement the waitgroup
				defer wg.Done()

				// Log status messages until the channel is closed, or a fatal error is retrieved.
				for M := range statusChan {
					CA.logger.Write([]byte(M.String()))
					if M.IsFatal {
						return
					}
				}

			}(wg)

			// If the plugin panics on Init, recover and stop the plugin.
			defer func(p ClientPlugin) {
				r := recover()
				if r != nil {
					if err := p.Stop(); err != nil {
						CA.logger.Write([]byte(err.Error() + "\n"))
					}
				}
			}(p)

			// Start the plugin.
			if err := p.Init(c, p.InputArguments...); err != nil {
				CA.logger.Write([]byte(err.Error() + "\n"))
				if err := p.Stop(); err != nil {
					CA.logger.Write([]byte(err.Error() + "\n"))
				}
			}

		}(CA.client, registeredPlugin, wg)
	}

	wg.Wait()

	// If all the registered plugins exit, then this agent is considered stopped.
	CA.stopped.Store(true)

	return nil
}

// Stop will cause ALL of the currentlyRunning Plugins to safely stop.
func (CA *ClientPluginAgent) Stop() error {

	CA.mux.Lock()
	defer CA.mux.Unlock()

	if dead, ok := CA.stopped.Load().(bool); ok {
		if dead {
			return nil
		}
	}
	CA.stopped.Store(true)

	errOccured := false

	wg := &sync.WaitGroup{}
	for _, p := range CA.RegisteredPlugins {
		wg.Add(1)

		go func(p ClientPlugin, wg *sync.WaitGroup) {

			// If the plugin exits, decrement the waitgroup
			defer wg.Done()

			if err := p.Stop(); err != nil {
				CA.logger.Write([]byte(err.Error() + "\n"))
				errOccured = true
			}

		}(p, wg)

	}

	wg.Wait()

	if errOccured {
		return errors.New("easytls agent error - error occured during client plugin shutdown")
	}

	return nil
}

// Wait for the plugin agent to stop safely.
func (CA *ClientPluginAgent) Wait() {
	for !CA.stopped.Load().(bool) {
		CA.logger.Write([]byte("easytls client plugin agent: Waiting to shut down...\n"))
		time.Sleep(time.Second)
	}
}

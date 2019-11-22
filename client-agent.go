package easytls

import (
	"errors"
	"io"
	"path"
	"path/filepath"
	"sort"
	"sync"
)

// ClientPluginAgent represents the a Plugin Manager agent to be used with a SimpleClient.
type ClientPluginAgent struct {
	client             *SimpleClient
	frameworkVersion   SemanticVersion
	registeredPlugins  []ClientPlugin
	logger             io.WriteCloser
	PluginSearchFolder string

	stopped bool
}

// NewClientAgent will create a new Client Plugin agent, ready to register plugins.
func NewClientAgent(Client *SimpleClient, Version SemanticVersion, PluginFolder string, Logger io.WriteCloser) (*ClientPluginAgent, error) {
	A := &ClientPluginAgent{
		client:             Client,
		frameworkVersion:   Version,
		PluginSearchFolder: PluginFolder,
		registeredPlugins:  []ClientPlugin{},
		logger:             Logger,
		stopped:            false,
	}

	return A, nil
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

	// Attempt to load all of the plugins.
	for _, f := range files {
		newPlugin, err := InitializeClientPlugin(f, CA.frameworkVersion)
		if err == nil {
			CA.registeredPlugins = append(CA.registeredPlugins, *newPlugin)
		}
	}

	return nil
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

	if len(CA.registeredPlugins) == 0 {
		return errors.New("easytls plugin error - Client Plugin Agent has 0 registered plugins")
	}

	wg := &sync.WaitGroup{}
	for _, registeredPlugin := range CA.registeredPlugins {
		wg.Add(1)

		// Start the plugin...
		go func(c *SimpleClient, p ClientPlugin, wg *sync.WaitGroup) {

			// Start logging plugin status messages.
			go func(wg *sync.WaitGroup) {

				// If the plugin exits, decrement the waitgroup
				defer wg.Done()

				// Extract the status channel
				statusChan, err := p.Status()

				// An error retrieving the status channel stops the logging.
				if err != nil {
					CA.logger.Write([]byte(err.Error()))
					return
				}

				// Log status messages until the channel is closed, or a fatal error is retrieved.
				for M := range statusChan {
					CA.logger.Write([]byte(M.String()))
					if M.IsFatal {
						return
					}
				}

			}(wg)

			// Start the plugin.
			if err := p.Init(c); err != nil {
				if err := p.Stop(); err != nil {
					CA.logger.Write([]byte(err.Error()))
				}
			}

		}(CA.client, registeredPlugin, wg)
	}

	wg.Wait()
	return nil
}

// Stop will cause ALL of the currentlyRunning Plugins to safely stop.
func (CA *ClientPluginAgent) Stop() error {
	defer func() { CA.stopped = true }()

	wg := &sync.WaitGroup{}
	for _, p := range CA.registeredPlugins {
		wg.Add(1)

		go func(p ClientPlugin, wg *sync.WaitGroup) {
			defer wg.Done()
			if err := p.Stop(); err != nil {
				CA.logger.Write([]byte(err.Error()))
			}
		}(p, wg)

	}

	wg.Wait()
	return nil
}

// Close down the plugin agent.
func (CA *ClientPluginAgent) Close() error {

	if !CA.stopped {
		if err := CA.Stop(); err != nil {
			CA.logger.Write([]byte(err.Error()))
		}
	}

	return CA.logger.Close()
}
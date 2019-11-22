package easytls

import (
	"errors"
	"io"
	"path"
	"path/filepath"
	"sort"
	"sync"
)

// ServerPluginAgent represents the a Plugin Manager agent to be used with a SimpleServer.
type ServerPluginAgent struct {
	frameworkVersion   SemanticVersion
	registeredPlugins  []ServerPlugin
	logger             io.WriteCloser
	PluginSearchFolder string

	stopped bool
}

// NewServerAgent will create a new Server Plugin agent, ready to register plugins.
func NewServerAgent(Version SemanticVersion, PluginFolder string, Logger io.WriteCloser) (*ServerPluginAgent, error) {
	A := &ServerPluginAgent{
		frameworkVersion:   Version,
		PluginSearchFolder: PluginFolder,
		registeredPlugins:  []ServerPlugin{},
		logger:             Logger,
		stopped:            false,
	}

	return A, nil
}

// RegisterPlugins will configure and register all of the plugins in the previously specified PluginFolder.  This will not start any of the plugins, but will only load the necessary symbols from them.
func (SA *ServerPluginAgent) RegisterPlugins() error {

	// Search for all plugins in the designated search folder...
	files, err := filepath.Glob(path.Join(SA.PluginSearchFolder, "*.so"))
	if err != nil {
		return err
	}

	// Sort all files alphabetically, to assert a standard import order
	sort.Strings(files)

	// Attempt to load all of the plugins.
	for _, f := range files {
		newPlugin, err := InitializeServerPlugin(f, SA.frameworkVersion)
		if err == nil {
			SA.registeredPlugins = append(SA.registeredPlugins, *newPlugin)
		}
	}

	return nil
}

// Run will start the ServerPlugin Agent, starting each of the registered plugins.
// blocking represents if the rest of the application should block on this SAll or not.
func (SA *ServerPluginAgent) Run(blocking bool) error {

	if blocking {
		return SA.run()
	}

	go SA.run()
	return nil
}

func (SA *ServerPluginAgent) run() error {

	if len(SA.registeredPlugins) == 0 {
		return errors.New("easytls plugin error - Server Plugin Agent has 0 registered plugins")
	}

	wg := &sync.WaitGroup{}
	for _, registeredPlugin := range SA.registeredPlugins {
		wg.Add(1)

		// Start the plugin...
		go func(p ServerPlugin, wg *sync.WaitGroup) {

			// Start logging plugin status messages.
			go func(wg *sync.WaitGroup) {

				// If the plugin exits, decrement the waitgroup
				defer wg.Done()

				// Extract the status channel
				statusChan, err := p.Status()

				// An error retrieving the status channel stops the logging.
				if err != nil {
					SA.logger.Write([]byte(err.Error()))
					return
				}

				// Log status messages until the channel is closed, or a fatal error is retrieved.
				for M := range statusChan {
					SA.logger.Write([]byte(M.String()))
					if M.IsFatal {
						return
					}
				}

			}(wg)

		}(registeredPlugin, wg)
	}

	wg.Wait()
	return nil
}

// Stop will SAuse ALL of the currentlyRunning Plugins to safely stop.
func (SA *ServerPluginAgent) Stop() error {
	defer func() { SA.stopped = true }()

	wg := &sync.WaitGroup{}
	for _, p := range SA.registeredPlugins {
		wg.Add(1)

		go func(p ServerPlugin, wg *sync.WaitGroup) {
			defer wg.Done()
			if err := p.Stop(); err != nil {
				SA.logger.Write([]byte(err.Error()))
			}
		}(p, wg)

	}

	wg.Wait()
	return nil
}

// Close down the plugin agent.
func (SA *ServerPluginAgent) Close() error {

	if !SA.stopped {
		if err := SA.Stop(); err != nil {
			SA.logger.Write([]byte(err.Error()))
		}
	}

	return SA.logger.Close()
}

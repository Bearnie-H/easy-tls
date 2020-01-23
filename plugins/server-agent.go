package plugins

import (
	"errors"
	"fmt"
	"io"
	"log"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ServerPluginAgent represents the a Plugin Manager agent to be used with a SimpleServer.
type ServerPluginAgent struct {
	frameworkVersion   SemanticVersion
	RegisteredPlugins  []ServerPlugin
	logger             io.WriteCloser
	PluginSearchFolder string

	// Protects the Stop and Close operations, making them idempotent.
	mux     sync.Mutex
	stopped atomic.Value
}

// NewServerAgent will create a new Server Plugin agent, ready to register plugins.
func NewServerAgent(PluginFolder string, Logger io.WriteCloser) (*ServerPluginAgent, error) {

	A := &ServerPluginAgent{
		frameworkVersion:   ServerFrameworkVersion,
		PluginSearchFolder: PluginFolder,
		RegisteredPlugins:  []ServerPlugin{},
		logger:             Logger,
		stopped:            atomic.Value{},
		mux:                sync.Mutex{},
	}

	A.stopped.Store(false)

	return A, nil
}

// GetPluginByName will return a pointer to the requested plugin.  This is typically used to provide input arguments for when the plugin is Initiated.
func (SA *ServerPluginAgent) GetPluginByName(Name string) (*ServerPlugin, error) {
	Name = strings.ToLower(Name)
	for index, p := range SA.RegisteredPlugins {
		name := strings.ToLower(p.Name())
		if strings.HasPrefix(name, Name) {
			return &(SA.RegisteredPlugins[index]), nil
		}
	}
	return nil, fmt.Errorf("easytls plugin error - Failed to find plugin %s", Name)
}

// StopPluginByName will attempt to stop a given plugin by name, if it exists
func (SA *ServerPluginAgent) StopPluginByName(Name string) error {

	p, err := SA.GetPluginByName(Name)
	if err != nil {
		return err
	}
	return p.Stop()
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

	var loadErrors error

	// Attempt to load all of the plugins.
	for _, f := range files {
		newPlugin, err := InitializeServerPlugin(f, SA.frameworkVersion)
		if err == nil {
			SA.RegisteredPlugins = append(SA.RegisteredPlugins, *newPlugin)
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

	if len(SA.RegisteredPlugins) == 0 {
		return errors.New("easytls plugin error - Server Plugin Agent has 0 registered plugins")
	}

	wg := &sync.WaitGroup{}
	for _, registeredPlugin := range SA.RegisteredPlugins {
		wg.Add(1)

		// Start the plugin...
		go func(p ServerPlugin, wg *sync.WaitGroup) {

			// If the plugin exits, decrement the waitgroup
			defer wg.Done()

			// Extract the status channel
			statusChan, err := p.Status()
			if err != nil {
				SA.logger.Write([]byte(err.Error() + "\n"))
				return
			}

			// Log status messages until the channel is closed, or a fatal error is retrieved.
			for M := range statusChan {
				SA.logger.Write([]byte(M.String()))
				if M.IsFatal {
					return
				}
			}

		}(registeredPlugin, wg)
	}

	wg.Wait()

	log.Printf("All plugins returned!")

	// Force this to wait for the Stop to fully complete.
	SA.Wait()

	log.Printf("Plugin agent successfully stopped!")

	return nil
}

// Stop will cause ALL of the currently Running Plugins to safely stop.
func (SA *ServerPluginAgent) Stop() error {

	SA.mux.Lock()
	defer SA.mux.Unlock()

	if dead, ok := SA.stopped.Load().(bool); ok {
		if dead {
			return nil
		}
	}
	defer SA.stopped.Store(true)

	errOccured := false

	wg := &sync.WaitGroup{}
	for _, p := range SA.RegisteredPlugins {
		wg.Add(1)

		go func(p ServerPlugin, wg *sync.WaitGroup) {

			// If the plugin exits, decrement the waitgroup
			defer wg.Done()

			SA.logger.Write([]byte(PluginStatus{Message: fmt.Sprintf("Stopping module [ %s ]...", p.Name())}.String()))
			defer SA.logger.Write([]byte(PluginStatus{Message: fmt.Sprintf("Stopped module [ %s ].", p.Name())}.String()))

			if err := p.Stop(); err != nil {
				SA.logger.Write([]byte(err.Error() + "\n"))
				errOccured = true
				return
			}

		}(p, wg)

	}

	wg.Wait()

	if errOccured {
		return errors.New("easytls agent error - error occured during server plugin shutdown")
	}

	return nil
}

// Wait for the plugin agent to stop safely.
func (SA *ServerPluginAgent) Wait() {
	for !SA.stopped.Load().(bool) {
		SA.logger.Write([]byte("easytls server plugin agent: Waiting to shut down...\n"))
		time.Sleep(time.Second)
	}
}

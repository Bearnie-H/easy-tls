package plugins

import (
	"errors"
	"log"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	easytls "github.com/Bearnie-H/easy-tls"
	"github.com/Bearnie-H/easy-tls/server"
)

var (
	// ErrNoModule indicates that no module exists for a given search
	ErrNoModule = errors.New("plugin agent error: No module found for key")
)

// Agent represets the Generic Plugin manager which oversees and acts on the loadable
// plugins. For specific Client and Server plugin managers, this must be extended
// with either ClientAgent or ServerAgent objects.
type Agent struct {

	// The version of the framework this plugin implements.
	// This is used to query the plugins to check for compatability.
	version SemanticVersion

	// The logger to write Status messages from the plugins to, as well
	// as agent-level messages and errors.
	logger *log.Logger

	// Protects loadedModules, the set of currently loaded but not
	// necessarily active modules.
	mu            *sync.Mutex
	loadedModules map[string]Module

	// moduleFolder is the folder to search in for modules to load.
	moduleFolder string

	// commandServer is the HTTP Server exposed over a Unix Socket
	// to allow for IPC to this agent to start, stop, and perform other
	// operations on the plugins available to it.
	commandServer     *server.SimpleServer
	commandServerSock string

	done chan struct{}
}

// NewAgent will create and return a new generic plugin agent.
func NewAgent(ModuleFolder string, logger *log.Logger) (*Agent, error) {

	var err error

	if logger == nil {
		logger = easytls.NewDefaultLogger()
	}

	A := &Agent{
		version:       GenericFrameworkVersion,
		logger:        logger,
		mu:            &sync.Mutex{},
		loadedModules: make(map[string]Module),
		done:          make(chan struct{}, 1),
	}

	if ModuleFolder, err = filepath.Abs(ModuleFolder); err != nil {
		return A, err
	}

	A.moduleFolder = ModuleFolder

	A.commandServer, err = newCommandServer(A)
	if err != nil {
		return A, err
	}

	// Load the modules from disk, putting this agent into a position where it could be started.
	if err := A.loadModules(); err != nil {
		return nil, err
	}

	return A, nil
}

func (A *Agent) loadModules() error {

	files, err := filepath.Glob(path.Join(A.moduleFolder, "*.so"))
	if err != nil {
		return err
	}

	for _, f := range files {

		P := A.NewGenericModule(f)

		if err := P.Load(); err != nil {
			return err
		}

		if err := A.registerModule(P); err != nil {
			return err
		}
	}

	return nil
}

// Logger returns the logger used by an Agent
func (A *Agent) Logger() *log.Logger { return A.logger }

// SetLogger will set the internal logger of the Agent to l
func (A *Agent) SetLogger(l *log.Logger) { A.logger = l }

// NewGenericModule will return a new GenericPlugin, which satisfies most, but not all
// of the Module interface.
// This GenericPlugin should be embedded into a greater object which fills in the final
// missing members of the Module interface.
func (A *Agent) NewGenericModule(Filename string) *GenericPlugin {

	p := &GenericPlugin{
		filename: Filename,
		agent:    A,
		done:     make(chan struct{}),
		args:     []interface{}{},
		logger:   A.logger,
		mu:       &sync.Mutex{},
		state:    stateNotLoaded,
		status:   nil,
		name:     nil,
		version:  nil,
		stop:     nil,
		init:     nil,
	}

	return p
}

// StartAll will run Start() for all of the included modules of the agent.
func (A *Agent) StartAll() error {

	wg := &sync.WaitGroup{}

	for _, M := range A.Modules() {
		wg.Add(1)
		go func(M Module, wg *sync.WaitGroup) {
			defer wg.Done()
			if err := M.Start(); err != nil {
				A.Logger().Printf("plugin agent error: Error occurred while starting module [ %s ] - %s", M.Name(), err)
			}
		}(M, wg)
	}

	wg.Wait()

	return nil
}

// GetByName will attempt to retrieve a module with a given name substring
// from the agent to be able to manipulate it.
func (A *Agent) GetByName(Substring string) (Module, error) {

	Name := A.canonicalizeName(Substring)
	if Name == "" {
		return nil, ErrNoModule
	}

	for _, M := range A.Modules() {
		if M != nil && M.Name() == Name {
			return M, nil
		}
	}

	return nil, ErrNoModule
}

// AddCommonArguments will add a set of arguments to all currently loaded plugins
func (A *Agent) AddCommonArguments(args ...interface{}) {
	for _, M := range A.Modules() {
		M.AddArguments(args...)
	}
}

// StopAll will run Stop() for all of the included modules of the agent.
func (A *Agent) StopAll() error {

	wg := &sync.WaitGroup{}

	for _, M := range A.Modules() {
		wg.Add(1)
		go func(M Module, wg *sync.WaitGroup) {
			defer wg.Done()
			if err := M.Stop(); err != nil {
				A.Logger().Printf("plugin agent error: Error occurred while stopping module [ %s ] - %s", M.Name(), err)
			}
		}(M, wg)
	}

	wg.Wait()

	return nil
}

// Close will close down an agent, stopping all plugins and releasing all resources
func (A *Agent) Close() error {

	select {
	case _, ok := <-A.Done():
		if !ok {
			return nil
		}
	default:
	}

	defer func() {
		A.Done() <- struct{}{}
		close(A.Done())
	}()

	var err error

	A.Logger().Printf("Stopping all modules...")
	if err = A.StopAll(); err != nil {
		A.Logger().Printf("plugin agent error: Error(s) occurred while stopping - %s", err)
	}

	A.Logger().Printf("Shutting down command server...")
	if err = A.commandServer.Shutdown(); err != nil {
		A.Logger().Printf("plugin agent error: Error occurred while shutting down command server - %s", err)
	}

	return err
}

// Done returns a channel which strobes when the agent is safely closed.
func (A *Agent) Done() chan struct{} {
	return A.done
}

// Wait will wait for all of the included modules to signal they are done before returning.
func (A *Agent) Wait() {
	// Wait for the agent to fully close
	<-A.Done()
}

// CanonicalizeName will attempt to canonicalize the given name snippet
// to the canonical name of the corresponding module. Returns an empty string
// if no match is found.
func (A *Agent) canonicalizeName(Name string) string {

	for _, M := range A.Modules() {
		if strings.HasPrefix(M.Name(), Name) {
			return M.Name()
		}
	}

	return ""
}

func (A *Agent) registerModule(M Module) error {

	A.mu.Lock()
	defer A.mu.Unlock()

	if A.loadedModules == nil {
		A.loadedModules = make(map[string]Module)
	}

	A.Logger().Printf("Adding module [ %s ]", M.Name())
	A.loadedModules[M.Name()] = M

	return nil
}

// Modules will return the set of known modules as an array in name
// sorted order, regardless of the underlying formatting.
func (A *Agent) Modules() []Module {

	M := []Module{}
	Names := []string{}

	A.mu.Lock()
	defer A.mu.Unlock()

	for n := range A.loadedModules {
		Names = append(Names, n)
	}

	sort.Strings(Names)

	for i := range Names {
		M = append(M, A.loadedModules[Names[i]])
	}

	return M
}

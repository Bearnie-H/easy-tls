package plugins

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/Bearnie-H/easy-tls/server"
	"github.com/gorilla/mux"
)

// ServerAgent represents a plugin manager for Server-type plugins
type ServerAgent struct {
	Agent

	// A reference to the Server this agent will serve on
	server *server.SimpleServer
	// A reference to the base Router to use when registering routes
	router *mux.Router
	// The base URL to root all loaded modules at.
	urlRoot string
}

// NewServerAgent will fully create and initialize a new ServerAgent.
// This will NOT start any plugins, but will put the Agent
// into a state where all available plugins could be started.
func NewServerAgent(ModuleFolder, URLRoot string, Server *server.SimpleServer) (A *ServerAgent, err error) {

	A = &ServerAgent{}

	// If no server is provided, spawn a default
	if Server == nil {
		A.server, err = server.NewServerHTTP()
		if err != nil {
			return nil, err
		}
	} else {
		A.server = Server
	}

	// Assert that the URLRoot ends with a slash and is not empty
	if URLRoot == "" {
		A.urlRoot = "/"
		URLRoot = "/"
	} else if !strings.HasSuffix(URLRoot, "/") {
		A.urlRoot = URLRoot + "/"
	}

	// Create a new subrouter for the plugins to be registered to,
	// with a special case that being added to "/" allows using the top-level router.
	if A.urlRoot == "/" {
		A.router = A.server.Router()
	} else {
		A.router = A.server.Router().PathPrefix(URLRoot).Subrouter()
	}

	// Create the new generic component of the agent.
	OtherServerActive := false
	A.Agent, err = newAgent(ServerFrameworkVersion, ModuleFolder, A.server.Logger())
	if err == ErrOtherServerActive {
		OtherServerActive = true
	} else if err != nil {
		return nil, err
	}

	// Load the modules from disk, putting this agent into a position where it could be started.
	if err := A.loadModules(); err != nil {
		return nil, err
	}

	if OtherServerActive {
		return A, ErrOtherServerActive
	}

	return A, nil
}

// NewServerModule will create a new Server Module
func (A *ServerAgent) NewServerModule(Filename string) (*ServerPlugin, error) {

	P := &ServerPlugin{
		GenericPlugin: *A.NewGenericModule(Filename),
		agent:         A,
		initHandlers:  nil,
		initSubrouter: nil,
	}

	return P, P.Load()
}

func (A *ServerAgent) loadModules() error {

	files, err := filepath.Glob(path.Join(A.moduleFolder, "*.so"))
	if err != nil {
		return err
	}

	for _, f := range files {

		P, err := A.NewServerModule(f)
		if err != nil {
			return err
		}

		if err := A.registerModule(P); err != nil {
			return err
		}
	}

	return nil
}

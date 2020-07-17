package plugins

import (
	"errors"
	"strings"
	"sync"

	"github.com/Bearnie-H/easy-tls/server"
	"github.com/gorilla/mux"
)

// ServerAgent represents a plugin manager for Server-type plugins
type ServerAgent struct {
	*Agent

	// A reference to the Server this agent will serve on
	server *server.SimpleServer

	// Protects router, preventing race conditions when adding routes
	routerLock *sync.Mutex
	// A reference to the base Router to use when registering routes
	router *mux.Router
	// The base URL to root all loaded modules at.
	urlRoot string
}

// NewServerAgent will fully create and initialize a new ServerAgent.
// This will NOT start any plugins, but will put the Agent
// into a state where all available plugins could be started.
func NewServerAgent(ModuleFolder, URLRoot string, Server *server.SimpleServer) (A *ServerAgent, err error) {

	A = &ServerAgent{
		routerLock: &sync.Mutex{},
	}

	// If no server is provided, spawn a default
	if Server == nil {
		A.server = server.NewServerHTTP()
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
		A.router = A.server.Router().PathPrefix(A.urlRoot).Subrouter()
	}

	// Create the new generic component of the agent.
	OtherServerActive := false
	A.Agent, err = NewAgent(ModuleFolder, A.server.Logger())
	if err == ErrOtherServerActive {
		OtherServerActive = true
	} else if err != nil {
		return nil, err
	}
	A.Agent.version = ServerFrameworkVersion

	// Load the modules from disk, putting this agent into a position where it could be started.
	if err := A.loadModules(); err != nil {
		return nil, err
	}

	if OtherServerActive {
		return A, ErrOtherServerActive
	}

	return A, nil
}

func (A *ServerAgent) loadModules() error {

	ServerModules := []*ServerPlugin{}

	for _, M := range A.Modules() {

		p, ok := M.(*GenericPlugin)
		if !ok {
			return errors.New("server plugin error: Invalid module type")
		}

		P := &ServerPlugin{
			GenericPlugin: *p,
			agent:         A,
			initHandlers:  nil,
			initSubrouter: nil,
		}

		if err := P.Load(); err != nil {
			return err
		}

		ServerModules = append(ServerModules, P)
	}

	A.loadedModules = nil

	for _, M := range ServerModules {
		A.registerModule(M)
	}

	return nil
}

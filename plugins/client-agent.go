package plugins

import (
	"path"
	"path/filepath"

	"github.com/Bearnie-H/easy-tls/client"
)

// ClientAgent represents a plugin manager for Server-type plugins
type ClientAgent struct {
	Agent

	// A reference to the Client this agent will pass to all plugins, along
	// with any set arguments
	client *client.SimpleClient
}

// NewClientAgent will fully create and initialize a new ClientAgent.
// This will NOT start any plugins, but will put the Agent
// into a state where all available plugins could be started.
func NewClientAgent(ModuleFolder string, Client *client.SimpleClient) (A *ClientAgent, err error) {

	A = &ClientAgent{}

	// If no server is provided, spawn a default
	if Client == nil {
		A.client, err = client.NewClientHTTP()
		if err != nil {
			return nil, err
		}
	} else {
		A.client = Client
	}

	// Create the new generic component of the agent.
	OtherServerActive := false
	A.Agent, err = newAgent(ServerFrameworkVersion, ModuleFolder, A.client.Logger())
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

// NewClientModule will create a new Client Module
func (A *ClientAgent) NewClientModule(Filename string) (*ClientPlugin, error) {

	P := &ClientPlugin{
		GenericPlugin: *A.NewGenericModule(Filename),
		agent:         A,
		init:          nil,
	}

	return P, P.Load()
}

func (A *ClientAgent) loadModules() error {

	files, err := filepath.Glob(path.Join(A.moduleFolder, "*.so"))
	if err != nil {
		return err
	}

	for _, f := range files {

		P, err := A.NewClientModule(f)
		if err != nil {
			return err
		}

		if err := A.registerModule(P); err != nil {
			return err
		}
	}

	return nil
}

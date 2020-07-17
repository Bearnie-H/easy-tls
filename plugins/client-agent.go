package plugins

import (
	"errors"

	"github.com/Bearnie-H/easy-tls/client"
)

// ClientAgent represents a plugin manager for Server-type plugins
type ClientAgent struct {
	*Agent

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
		A.client = client.NewClientHTTP()
	} else {
		A.client = Client
	}

	// Create the new generic component of the agent.
	OtherServerActive := false
	A.Agent, err = NewAgent(ModuleFolder, A.client.Logger())
	if err == ErrOtherServerActive {
		OtherServerActive = true
	} else if err != nil {
		return nil, err
	}
	A.Agent.version = ClientFrameworkVersion

	// Load the modules from disk, putting this agent into a position where it could be started.
	if err := A.loadModules(); err != nil {
		return nil, err
	}

	if OtherServerActive {
		return A, ErrOtherServerActive
	}

	return A, nil
}

func (A *ClientAgent) loadModules() error {

	ClientModules := []*ClientPlugin{}

	for _, M := range A.Modules() {

		p, ok := M.(*GenericPlugin)
		if !ok {
			return errors.New("client plugin error: Invalid module type")
		}

		P := &ClientPlugin{
			GenericPlugin: *p,
			agent:         A,
			init:          nil,
		}

		if err := P.Load(); err != nil {
			return err
		}

		ClientModules = append(ClientModules, P)
	}

	A.loadedModules = nil

	for _, M := range ClientModules {
		A.registerModule(M)
	}

	return nil
}

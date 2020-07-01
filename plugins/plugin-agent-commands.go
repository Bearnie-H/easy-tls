package plugins

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/Bearnie-H/easy-tls/client"
)

type command struct {
	action string
	name   string
}

func newCommand(args ...string) (*command, []string, error) {

	if len(args) < 2 {
		return nil, nil, errors.New("plugin command error: Not enough arguments to build command")
	}

	c := &command{
		action: args[0],
		name:   args[1],
	}

	args = args[2:]

	return c, args, nil
}

func (c *command) url(Agent *Agent) *url.URL {

	if c.name != "" {
		c.name = "/" + c.name
	}

	return &url.URL{
		Scheme: "http",
		Host:   "unix",
		Path:   fmt.Sprintf("/%s%s", c.action, c.name),
	}
}

func (c *command) do(C *client.SimpleClient, A *Agent) error {

	resp, err := C.Get(c.url(A), nil)
	if err != nil && err != client.ErrInvalidStatusCode {
		A.Logger().Printf("plugin command error: Error occured while submitting command [ %+v ] - %s", c, err)
		return err
	}

	Contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		A.Logger().Printf("plugin command error: Error occured while reading response body - %s", err)
		return err
	}

	for _, s := range strings.Split(string(Contents), "\n") {
		if s != "" {
			A.Logger().Print(s)
		}
	}
	resp.Body.Close()

	return nil
}

// SendCommands is the top-level function to send a command to the command-server
// of another plugin agent, to cause it to perform actions on the set of plugins
// it has. These actions exactly correspond to the set of methods within the
// Module interface.
//
// A given command is composed of 2 strings:
//	Action
//	Name
//
// The receiving handler will attempt to perform the given action
// on the plugin who matches the Name. This can be just a prefix
// of the full name. The snippet provided will always match in
// lexicographic order, so if one module is a prefix of another, and
// a sub-prefix is given, the shortest name will match first.
func (A *Agent) SendCommands(Args ...string) {

	A.Logger().Println("Existing plugin agent active on socket, attempting to send commands.")

	if Args == nil || len(Args) == 0 {
		A.Logger().Println("No arguments provided, exiting.")
		return
	}

	// Create a new Client, with a customized dialer to communicate over the Unix socket.
	C := client.NewClient(&http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, addr string) (net.Conn, error) {
				dialer := net.Dialer{}
				return dialer.DialContext(ctx, "unix", A.commandServerSock)
			},
		},
	})

	// Share logging.
	C.SetLogger(A.Logger())

	A.sendCommands(C, Args...)
}

func (A *Agent) sendCommands(C *client.SimpleClient, args ...string) {

	c := command{
		action: args[0],
	}

	if len(args) == 1 {
		c.do(C, A)
	} else {
		for _, name := range args[1:] {
			c.name = name
			c.do(C, A)
		}
	}
}

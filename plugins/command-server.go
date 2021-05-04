package plugins

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path"

	"github.com/Bearnie-H/easy-tls/client"
	"github.com/Bearnie-H/easy-tls/server"
)

var (
	// ErrOtherServerActive indicates that the server cannot activate, since there is already a server
	// active on the bind address it uses.
	ErrOtherServerActive error = errors.New("plugin command server error: Failed to initialize, bind address already active")
)

// newCommandServer will return a new, fully configured Command server to the plugin agent.
func newCommandServer(Agent *Agent) (*server.SimpleServer, error) {

	Agent.commandServerSock = formatSocketName(Agent.moduleFolder)

	// Check if there is already a socket open with this name
	_, err := os.Stat(Agent.commandServerSock)
	if err == nil {

		// Check if someone is listening on the other end.
		if Agent.commandServerActive() {
			Agent.Logger().Printf("Plugin Agent socket already active.")
			return nil, ErrOtherServerActive
		}
		if err := os.Remove(Agent.commandServerSock); err != nil {
			return nil, err
		}
	}

	// Spawn a new listener to listen on the unix domain socket
	L, err := newCommandListener(Agent.commandServerSock)
	if err != nil {
		return nil, err
	}

	// Create the server
	Agent.Logger().Printf("Creating plugin command server at [ %s ]", L.Addr().String())
	S := server.NewServerHTTP(L.Addr().String())

	// Add in the dedicated handlers to perform actions on the plugins loaded by the agent
	S.AddHandlers(S.Router(), formatCommandHandlers(Agent)...)

	Agent.Logger().Printf("Serving command server at [ %s ]", S.Addr())

	// Serve traffic on the listener
	go func(S *server.SimpleServer, L *net.UnixListener) {
		S.Serve(L)
	}(S, L)

	return S, nil
}

func (A *Agent) commandServerActive() bool {

	// Create a new Client, with a customized dialer to communicate over the Unix socket.
	C := client.NewClient(&http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, addr string) (net.Conn, error) {
				dialer := net.Dialer{}
				return dialer.DialContext(ctx, "unix", A.commandServerSock)
			},
		},
	})

	// Don't log any of these intermediate steps
	l := A.Logger()

	discardLogger := log.New(ioutil.Discard, "", 0)
	C.SetLogger(discardLogger)
	A.SetLogger(discardLogger)

	defer A.SetLogger(l)

	// Create the simplest request to send and try to get a response
	c := command{action: "help", name: ""}

	if err := c.do(C, A); err != nil {
		return false
	}

	return true
}

func newCommandListener(SocketName string) (*net.UnixListener, error) {

	l, err := net.ListenUnix(
		"unix",
		&net.UnixAddr{
			Name: SocketName,
			Net:  "unix",
		},
	)

	return l, err
}

// formatSocketName will generate a unique valid name for a Unix socket to listen
// on and serve the command server.
func formatSocketName(Seed string) string {
	n := sha1.Sum([]byte(Seed))
	return path.Join("/tmp", hex.EncodeToString(n[:])+".sock")
}

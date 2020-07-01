package plugins

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"net"
	"os"
	"path"
	"runtime"
	"syscall"

	"github.com/Bearnie-H/easy-tls/server"
)

var (
	// ErrOtherServerActive indicates that the server cannot activate, since there is already a server
	// active on the bind address it uses.
	ErrOtherServerActive error = errors.New("plugin command server error: Failed to initalize, bind address already active")
)

// newCommandServer will return a new, fully configured Command server to the plugin agent.
func newCommandServer(Agent *Agent) (*server.SimpleServer, error) {

	Agent.commandServerSock = formatSocketName(Agent.moduleFolder)

	// Spawn a new listener to listen on the unix domain socket
	L, err := newCommandListener(Agent.commandServerSock)
	if err != nil {
		return nil, err
	}

	// Create the server
	Agent.Logger().Printf("Creating plugin command server at [ %s ]", L.Addr().String())
	S, err := server.NewServerHTTP(L.Addr().String())
	if err != nil {
		return nil, err
	}

	// Add in the dedicated handlers to perform actions on the plugins loaded by the agent
	S.AddHandlers(S.Router(), formatCommandHandlers(Agent)...)

	// Serve traffic on the listener
	S.Serve(L, true)

	Agent.Logger().Printf("Serving command server at [ %s ]", S.Addr())

	return S, nil
}

func newCommandListener(SocketName string) (*net.UnixListener, error) {

	l, err := net.ListenUnix(
		"unix",
		&net.UnixAddr{
			Name: SocketName,
			Net:  "unix",
		},
	)

	// Check if the error fails because the socket is already bound...
	if isErrorAddressAlreadyInUse(err) {
		err = ErrOtherServerActive
	}

	return l, err
}

// formatSocketName will generate a unique valid name for a Unix socket to listen
// on and serve the command server.
func formatSocketName(Seed string) string {
	n := sha1.Sum([]byte(Seed))
	return path.Join("/tmp", hex.EncodeToString(n[:])+".sock")
}

func isErrorAddressAlreadyInUse(err error) bool {

	errOpError, ok := err.(*net.OpError)
	if !ok {
		return false
	}

	errSyscallError, ok := errOpError.Err.(*os.SyscallError)
	if !ok {
		return false
	}

	errErrno, ok := errSyscallError.Err.(syscall.Errno)
	if !ok {
		return false
	}

	if errErrno == syscall.EADDRINUSE {
		return true
	}

	const WSAEADDRINUSE = 10048
	if runtime.GOOS == "windows" && errErrno == WSAEADDRINUSE {
		return true
	}

	return false
}

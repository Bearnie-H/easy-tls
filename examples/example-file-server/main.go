package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/Bearnie-H/easy-tls/server"
	"github.com/Bearnie-H/easy-tls/server/fileserver"
)

const (
	// DefaultServerAddress is the default address on which the HTTP server will be served.
	DefaultServerAddress = 8080

	// DefaultServeBase is the default directory served by this application
	DefaultServeBase = "./"

	// DefaultURLBase is the base URL to serve from.
	DefaultURLBase = "/"
)

// Command-line flags
var (
	Addr     = flag.Uint("port", DefaultServerAddress, "The TCP port to serve the file server on.")
	ServeDir = flag.String("folder", DefaultServeBase, "The base directory to serve files from.")
	URLRoot  = flag.String("url", DefaultURLBase, "The base URL to serve the file server from.")
)

func main() {
	flag.Parse()
	var err error

	// Create a new HTTP Server, which will listen on all interfaces.
	Server := server.NewServerHTTP(fmt.Sprintf(":%d", *Addr))

	*ServeDir, err = filepath.Abs(*ServeDir)
	if err != nil {
		panic(err)
	}

	log.Printf("Serving directory [ %s ] at URL [ %s ].", *ServeDir, *URLRoot)

	// Add some middlewares as an example
	Server.AddMiddlewares(server.MiddlewareDefaultLogger(Server.Logger()))
	Server.AddMiddlewares(server.MiddlewareLimitMaxConnections(200, time.Minute*15, nil))
	Server.AddMiddlewares(server.MiddlewareLimitConnectionRate(time.Millisecond*10, time.Minute*15, nil))

	// Add routes
	Handlers, err := fileserver.Handlers(*URLRoot, *ServeDir, false, Server.Logger())
	if err != nil {
		panic(err)
	}
	Server.AddHandlers(Server.Router(), Handlers...)

	// Set the server-side timeouts
	Server.SetTimeouts(time.Hour, time.Second*15, time.Hour, time.Second*5, 0)

	// Set up a go-routine to allow the application to safely shut down.
	initSafeShutdown(Server)

	// Start up the server, and block until it is closed
	if err := Server.ListenAndServe(); err != nil {
		log.Println(err)
	}
}

func initSafeShutdown(Server *server.SimpleServer) {
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, os.Kill)
	go doSafeShutdown(sigChan, Server)
}

func doSafeShutdown(C chan os.Signal, Server *server.SimpleServer) {

	// Wait on a signal
	<-C
	log.Println("Shutting down EasyTLS Server...")
	defer log.Println("Shut down EasyTLS Server!")

	// Close and stop the Server
	if err := Server.Shutdown(); err != nil {
		log.Println(err)
	}
}

// This package implements a minimalist example of a Server-Side Framework, using the EasyTLS library.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/Bearnie-H/easy-tls/plugins"
	"github.com/Bearnie-H/easy-tls/proxy"
	"github.com/Bearnie-H/easy-tls/server"
)

// Example Constants, set these based on your application needs
const (
	ModuleFolder     = "./active-modules"
	ServerAddress    = ":8080"
	RoutingRulesFile = "./EasyTLS-Proxy.rules"
)

func main() {

	flag.Parse()

	// Create a new HTTP Server, which will listen on port 8080 on all interfaces.
	Server := server.NewServerHTTP(ServerAddress)

	// Create a new plugin agent to load the modules into.
	// If this fails with ErrOtherServerActive, this indicates there's
	// already an identical plugin agent up and running, so this will instead attempt
	// to send commands to it, using any remaining command-line arguments to
	// build the commands to send.
	Agent, err := plugins.NewServerAgent(ModuleFolder, "/", Server)
	if err == plugins.ErrOtherServerActive {
		os.Exit(0)
	} else if err != nil {
		panic(err)
	}

	// Add some middlewares as an example
	Server.AddMiddlewares(server.MiddlewareLimitConnectionRate(time.Millisecond*10, time.Minute*15, Server.Logger()))
	Server.AddMiddlewares(server.MiddlewareLimitMaxConnections(200, time.Minute*15, Server.Logger()))
	Server.AddMiddlewares(server.MiddlewareDefaultLogger(Server.Logger()))

	// Set up the server so that any routes which are not found are checked against a routing table file, allowing this server to proxy requests it cannot serve itself, but has been configured to proxy for.
	proxy.NotFoundHandlerProxyOverride(Server, nil, proxy.LiveFileRouter(RoutingRulesFile), nil)

	// Set the server-side timeouts
	Server.SetTimeouts(time.Hour, time.Second*15, time.Hour, time.Second*5)

	// Start all the modules, loading all the routes and beginning status logging.
	if err := Agent.StartAll(); err != nil {
		panic(err)
	}

	// Set up a go-routine to allow the application to safely shut down.
	initSafeShutdown(Agent, Server)

	// Start up the server, and block until it is closed
	if err := Server.ListenAndServe(); err != nil {
		log.Println(err)
	}

	Agent.Wait()
}

func initSafeShutdown(A *plugins.ServerAgent, S ...*server.SimpleServer) {
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, os.Kill)
	go doSafeShutdown(sigChan, A, S...)
}

func doSafeShutdown(C chan os.Signal, A *plugins.ServerAgent, S ...*server.SimpleServer) {

	// Wait on a signal
	<-C
	log.Println("Shutting down EasyTLS Server framework...")

	Done := make(chan struct{}, 1)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	go func(Done chan<- struct{}) {

		// Close and stop the Server(s)
		for _, srv := range S {
			if err := srv.Shutdown(); err != nil {
				log.Println(err)
			}
		}

		// Close and stop the Plugin Agent
		if err := A.Close(); err != nil {
			log.Println(err)
		}
	}(Done)

	select {
	case <-Done:
		log.Println("Successfully shut down EasyTLS Server Framework!")
	case <-ctx.Done():
		log.Println("Failed to shut down EasyTLS Server Framework before timeout!")
		os.Exit(1)
	}
}

// This package implements a minimalist example of a Server-Side Framework, using the EasyTLS library.
package main

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/Bearnie-H/easy-tls/plugins"
	"github.com/Bearnie-H/easy-tls/server"
)

func main() {

	// Create a new plugin agent, reading plugins from ./active-modules and logging to stdout
	Agent, err := plugins.NewServerAgent("./active-modules", os.Stdout)
	if err != nil {
		panic(err)
	}

	if err := Agent.RegisterPlugins(); err != nil {
		panic(err)
	}

	// Start status logging of the plugins.
	Agent.Run(false)

	// Create a new HTTP Server, which will listen on port 8080 on all interfaces.
	Server, err := server.NewServerHTTP(":8080")
	if err != nil {
		panic(err)
	}

	// Create a new default router, which will have routes added from the plugins.
	router := server.NewDefaultRouter()

	// Walk the set of registered plugins, adding the routes from each to the router.
	for _, p := range Agent.RegisteredPlugins {
		routes, err := p.Init()
		if err != nil {
			panic(err)
		}
		server.AddHandlers(true, Server, router, routes...)
	}

	// Add some middlewares as an example
	server.AddMiddlewares(router, server.MiddlewareLimitConnectionRate(time.Millisecond*10, time.Minute*15, true))
	server.AddMiddlewares(router, server.MiddlewareLimitMaxConnections(200, time.Minute*15, true))

	// Add in a route to display a route guide
	Server.EnableAboutHandler(router)

	// Register the router, to allow the server to actually serve the routes defined.
	Server.RegisterRouter(router)

	// Set the server-side timeouts
	Server.SetTimeouts(time.Hour, time.Second*15, time.Hour, time.Second*5)

	// Set up a go-routine to allow the application to safely shut down.
	initSafeShutdown(Server, Agent)

	// Start up the server, and block until it is closed
	if err := Server.ListenAndServe(); err != nil {
		log.Println(err)
	}
}

func initSafeShutdown(S *server.SimpleServer, A *plugins.ServerPluginAgent) {
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, os.Kill)
	go doSafeShutdown(sigChan, S, A)
}

func doSafeShutdown(C chan os.Signal, S *server.SimpleServer, A *plugins.ServerPluginAgent) {

	// Wait on a signal
	<-C
	log.Println("Shutting down...")

	// Close and stop the Plugin Agent
	if err := A.Close(); err != nil {
		log.Println(err)
	}

	// Close and stop the Server
	if err := S.Shutdown(); err != nil {
		log.Println(err)
	}
}

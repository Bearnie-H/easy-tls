// This package implements a minimalist example of a Server-Side Framework, using the EasyTLS library.
package main

import (
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

	// Create a new plugin agent, reading plugins from ./active-modules and logging to stdout
	Agent, err := plugins.NewServerAgent(ModuleFolder, os.Stdout)
	if err != nil {
		panic(err)
	}

	if err := Agent.RegisterPlugins(); err != nil {
		log.Println(err)
	}

	// Start status logging of the plugins.
	Agent.Run(false)

	// Create a new HTTP Server, which will listen on port 8080 on all interfaces.
	Server, err := server.NewServerHTTP(ServerAddress)
	if err != nil {
		panic(err)
	}

	// Create a new default router, which will have routes added from the plugins.
	router := server.NewDefaultRouter()

	// Walk the set of registered plugins, adding the routes from each to the router.
	for _, p := range Agent.RegisteredPlugins {
		routes, err := p.Init()
		if err != nil {
			log.Println(err)
			continue
		}
		server.AddHandlers(true, Server, router, routes...)
	}

	// Add some middlewares as an example
	server.AddMiddlewares(router, server.MiddlewareLimitConnectionRate(time.Millisecond*10, time.Minute*15, true))
	server.AddMiddlewares(router, server.MiddlewareLimitMaxConnections(200, time.Minute*15, true))

	// Add in a route to display a route guide
	Server.EnableAboutHandler(router)

	// Set up the server so that any routes which are not found are checked against a routing table file, allowing this server to proxy requests it cannot serve itself, but has been configured to proxy for.
	proxy.NotFoundHandlerProxyOverride(router, proxy.LiveFileRouter(RoutingRulesFile), true)

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

	Agent.Wait()
}

func initSafeShutdown(S *server.SimpleServer, A *plugins.ServerPluginAgent) {
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, os.Kill)
	go doSafeShutdown(sigChan, A, S)
}

func doSafeShutdown(C chan os.Signal, A *plugins.ServerPluginAgent, S ...*server.SimpleServer) {

	// Wait on a signal
	<-C
	log.Println("Shutting down EasyTLS Server framework...")
	defer log.Println("Shut down EasyTLS Server framework!")

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
}

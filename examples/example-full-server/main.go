package main

import (
	"database/sql"
	"log"
	"os"
	"os/signal"

	"github.com/Bearnie-H/easy-tls/plugins"
	"github.com/Bearnie-H/easy-tls/proxy"
	"github.com/Bearnie-H/easy-tls/server"
	"github.com/Bearnie-H/easy-tls/server/fileserver"
)

var (
	done chan struct{} = make(chan struct{})
)

func initSafeShutdown(Agent *plugins.ServerAgent, Servers ...*server.SimpleServer) {
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, os.Kill)
	go doSafeShutdown(sigChan, Agent, Servers...)
}

func doSafeShutdown(C chan os.Signal, Agent *plugins.ServerAgent, Servers ...*server.SimpleServer) {

	// Wait on a signal
	Sig := <-C
	log.Printf("Received Signal [ %s ].", Sig.String())
	log.Println("Shutting down application...")
	defer log.Println("Shut down application!")

	// Close and stop the plugin agent
	if err := Agent.Close(); err != nil {
		log.Println(err)
	}

	// Close and stop any servers
	for _, S := range Servers {
		if err := S.Shutdown(); err != nil {
			log.Println(err)
		}
	}
	done <- struct{}{}
}

// This example provides a purposefully "busy" server as a full
// demonstration of all the functionalities provided by this library for
// configuring and setting up servers. All of these various handlers and
// "sub-applications" can all be served from the same server as long as
// the URL trees are all disjoint.
func main() {

	// Create one server to listen on all interfaces on port 8080
	S, err := server.NewServerHTTP()
	if err != nil {
		panic(err)
	}

	// Create a plugin agent to let this application load plugin modules
	Agent, err := plugins.NewServerAgent("./plugins", "/", S)
	if err != nil {
		panic(err)
	}

	// Set up a go-routine to allow the application to safely shut down.
	initSafeShutdown(Agent, S)

	// Add a file-server to serve the full "/tmp" directory based off URL "/tmp"
	S.AddSubrouter(S.Router(), "/tmp", fileserver.Handlers("/tmp", "/tmp", false, S.Logger())...)

	// Add a second file-server to serve "/var/log" from "/log"
	S.AddSubrouter(S.Router(), "/log", fileserver.Handlers("/log", "/var/log", false, S.Logger())...)

	// Add a different file-server to the second server...
	S.AddSubrouter(S.Router(), "/home", fileserver.Handlers("/home", "/home/", false, S.Logger())...)

	// Set up a proxy server, with forwarding rules defined in a live-editable file
	// to listen and proxy anything coming in on /proxy
	proxy.ConfigureReverseProxy(
		S,
		nil,
		S.Logger(),
		proxy.LiveFileRouter("./EasyTLS-Proxy.rules"),
		"/proxy",
	)

	// Serve a Single Page Application (Angular, React, etc...)
	// at "/ui". Note, this must match what the SPA considers the Base URL
	// to ensure relative links function correctly.
	if err := S.RegisterSPAHandler("/ui-1", "./static/app"); err != nil {
		panic(err)
	}

	// Serve a second SPA from the main server
	if err := S.RegisterSPAHandler("/ui-2", "./static/app2"); err != nil {
		panic(err)
	}

	// Open a connection to a database, to give access to the plugins
	Handle, err := sql.Open("", "")
	if err != nil {
		// The above call won't work since it has no arguments, so just ignore the error
		// as this is just an example.
		// panic(err)
	}

	// Give every plugin a copy of this database handle
	Agent.AddCommonArguments(Handle)

	// Log all incoming connections
	S.AddMiddlewares(server.MiddlewareDefaultLogger(S.Logger()))

	if err := Agent.StartAll(); err != nil {
		panic(err)
	}

	// Start the server
	if err := S.ListenAndServe(true); err != nil {
		log.Println(err)
	}

	<-done
}

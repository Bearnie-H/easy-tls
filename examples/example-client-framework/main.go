package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/Bearnie-H/easy-tls/plugins"
)

func main() {

	// Create a Plugin Agent, which will create a default HTTP Client, to use modules found in ./active-modules, and to log all output to STDOUT.
	Agent, err := plugins.NewClientAgent(nil, "./active-modules", os.Stdout)
	if err != nil {
		panic(err)
	}

	// Set up a go-routine to allow the application to safely shut down.
	initSafeShutdown(Agent)

	// Start the plugins, and block until they all have stopped.
	Agent.Run(true)

	Agent.Wait()
}

func initSafeShutdown(A *plugins.ClientPluginAgent) {
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, os.Kill)
	go doSafeShutdown(sigChan, A)
}

func doSafeShutdown(C chan os.Signal, A *plugins.ClientPluginAgent) {

	// Wait on a signal
	<-C
	log.Println("Shutting down...")

	// Close and stop the Plugin Agent
	if err := A.Close(); err != nil {
		log.Println(err)
	}
}

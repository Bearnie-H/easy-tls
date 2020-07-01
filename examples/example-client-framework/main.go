package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/Bearnie-H/easy-tls/plugins"
)

func main() {

	// Create a Plugin Agent, which will create a default HTTP Client, to use modules found in ./active-modules, and to log all output to STDOUT.
	Agent, err := plugins.NewClientAgent("./active-modules", nil)
	if err == plugins.ErrOtherServerActive {
		os.Exit(0)
	} else if err != nil {
		panic(err)
	}

	Agent.StartAll()

	// Set up a go-routine to allow the application to safely shut down.
	initSafeShutdown(Agent)

	Agent.Wait()
}

func initSafeShutdown(A *plugins.ClientAgent) {
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, os.Kill)
	go doSafeShutdown(sigChan, A)
}

func doSafeShutdown(C chan os.Signal, A *plugins.ClientAgent) {

	// Wait on a signal
	<-C
	log.Println("Shutting down EasyTLS Client Framework...")
	defer log.Println("Shut down EasyTLS Client Framework!")

	// Close and stop the Plugin Agent
	if err := A.Close(); err != nil {
		log.Println(err)
	}
}

package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/Bearnie-H/easy-tls/plugins"
)

func main() {
	flag.Parse()

	// Create a Plugin Agent, which will create a default HTTP Client, to use modules found in ./active-modules, and to log all output to STDOUT.
	Agent, err := plugins.NewAgent("./active-modules", nil)
	if err == plugins.ErrOtherServerActive {
		Agent.SendCommands(flag.Args()...)
		os.Exit(0)
	} else if err != nil {
		panic(err)
	}

	Agent.StartAll()

	// Set up a go-routine to allow the application to safely shut down.
	initSafeShutdown(Agent)

	Agent.Wait()
}

func initSafeShutdown(A *plugins.Agent) {
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, os.Kill)
	go doSafeShutdown(sigChan, A)
}

func doSafeShutdown(C chan os.Signal, A *plugins.Agent) {

	// Wait on a signal
	<-C
	log.Println("Shutting down EasyTLS Framework...")
	defer log.Println("Shut down EasyTLS Framework!")

	// Close and stop the Plugin Agent
	if err := A.Close(); err != nil {
		log.Println(err)
	}
}

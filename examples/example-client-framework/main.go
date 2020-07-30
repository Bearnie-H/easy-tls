package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/Bearnie-H/easy-tls/plugins"
)

func main() {
	flag.Parse()

	// Create a Plugin Agent, which will create a default HTTP Client, to use modules found in ./active-modules, and to log all output to STDOUT.
	Agent, err := plugins.NewClientAgent("./active-modules", nil)
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

func initSafeShutdown(A *plugins.ClientAgent) {
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, os.Kill)
	go doSafeShutdown(sigChan, A)
}

func doSafeShutdown(C chan os.Signal, A *plugins.ClientAgent) {

	// Wait on a signal
	<-C
	log.Println("Shutting down EasyTLS Client Framework...")

	Done := make(chan struct{}, 1)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	go func(Done chan<- struct{}) {

		// Close and stop the Plugin Agent
		if err := A.Close(); err != nil {
			log.Println(err)
		}
	}(Done)

	select {
	case <-Done:
		log.Println("Successfully shut down EasyTLS Client Framework!")
	case <-ctx.Done():
		log.Println("Failed to shut down EasyTLS Client Framework before timeout!")
		os.Exit(1)
	}
}

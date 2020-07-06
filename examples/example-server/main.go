package main

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/Bearnie-H/easy-tls/server"
)

//	ServerAddress is the default address on which the HTTP server will be served.
const (
	ServerAddress = ":8080"
)

func main() {

	// Create a new HTTP Server, which will listen on port 8080 on all interfaces.
	Server := server.NewServerHTTP(ServerAddress)

	// Add some middlewares as an example
	Server.AddMiddlewares(server.MiddlewareLimitConnectionRate(time.Millisecond*10, time.Minute*15, Server.Logger()))
	Server.AddMiddlewares(server.MiddlewareLimitMaxConnections(200, time.Minute*15, Server.Logger()))
	Server.AddMiddlewares(server.MiddlewareDefaultLogger(Server.Logger()))

	// Add routes
	addRoutes(Server)

	// Set the server-side timeouts
	Server.SetTimeouts(time.Hour, time.Second*15, time.Hour, time.Second*5)

	// Set up a go-routine to allow the application to safely shut down.
	initSafeShutdown(Server)

	// Start up the server, and block until it is closed
	if err := Server.ListenAndServe(); err != nil {
		log.Println(err)
	}
}

func addRoutes(Server *server.SimpleServer) {

	// ...

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

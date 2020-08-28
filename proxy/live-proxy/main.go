package main

import (
	"flag"
	"fmt"

	"github.com/Bearnie-H/easy-tls/proxy"
	"github.com/Bearnie-H/easy-tls/server"
)

// Define the command-line arguments
var (
	InterfaceFlag = flag.String("addr", "", "The interface to serve HTTP on.")
	PortFlag      = flag.Int("port", 8080, "The port to serve HTTP on.")
	RulesFilename = flag.String("file", "EasyTLS-Proxy.rules", "The filename of the EasyTLS Proxy Rules file to work with.")
)

func main() {
	flag.Parse()

	// Configure the proxy, start listening and serving, and if any errors happen, panic to report them.
	if err := proxy.ConfigureReverseProxy(
		server.NewServerHTTP(fmt.Sprintf("%s:%d", *InterfaceFlag, *PortFlag)),
		nil,
		nil,
		proxy.LiveFileRouter(*RulesFilename),
		"",
	).ListenAndServe(); err != nil {
		panic(err)
	}

}

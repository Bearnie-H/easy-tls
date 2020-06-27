package main

import (
	"flag"
	"fmt"

	"github.com/Bearnie-H/easy-tls/proxy"
	"github.com/Bearnie-H/easy-tls/server"
)

// Define the command-line arguments
var (
	AddressFlag   = flag.String("addr", "", "The address to serve HTTP on.")
	PortFlag      = flag.Int("port", 8080, "The port to serve HTTP on.")
	RulesFilename = flag.String("file", "EasyTLS-Proxy.rules", "The filename of the EasyTLS Proxy Rules file to work with.")
)

func main() {
	flag.Parse()

	s, err := server.NewServerHTTP(fmt.Sprintf("%s:%d", *AddressFlag, *PortFlag))
	if err != nil {
		panic(err)
	}

	proxy.ConfigureReverseProxy(s, nil, s.Logger(), proxy.LiveFileRouter(*RulesFilename), "/")

	if err := s.ListenAndServe(); err != nil {
		panic(err)
	}
}

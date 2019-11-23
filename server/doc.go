// Package server implements a minor extension of the standard http.Server struct.
//
// The primary extensions of this implementation over the standard library are:
//
//	1) Integration with the base EasyTLS package, to provide a standardized and simpler interface for configuring and utilizing TLS encryption on an HTTP Server.
//
//	2) Slightly easier interface for starting and stopping the server, registering routes and middlewares, and a small library of included middlewares for common use-cases.
//
//	3) Plugin integration via the easy-tls/plugins package.  Client/Server application design can be simplified or broken down from monoliths with an easy enough plugin system, which this is intended to provide.  The http.Handler registration process has been simplified in such a way as to allow programmatic registration of Routes by the use of SimpleHandlers.  A Server module can easily expose the set of routes it services as an array of SimpleHandlers, which can be iterated over and registered.
//
package server

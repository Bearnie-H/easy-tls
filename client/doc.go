// Package client implements a minor extension of the default http.Client.
//
// The primary extensions of this implementation over the standard library are:
//
//	1) Integration with the "easytls" package, to provide a standardized and
//	simpler interface for configuring and utilizing TLS encryption on an HTTP Client.
//
//	This also provides the capability to downgrade from HTTPS to HTTP if the
//	server responds with HTTP, and to allow the client to reset back to HTTPS
//	after the request.
//
//	2) Explicit exposure and availability of all standard HTTP methods with a
//	standardized calling convention. By default, the standard HTTP package only
//	provides Get() and Post() as dedicated functions, and defers to Do() for
//	all other request types. This works, but becomes unhelpful when interacting
//	with more complex services that use the other methods.
//
//	3) Plugin integration via the easy-tls/plugins package. Client/Server
//	application design can be broken down with an easy enough plugin system,
//	which this is intended to provide. The underlying http.Client is safe for
//	concurrent use, therefore the plugin architecture assumes that a single
//	SimpleClient instance can be passed in to any number of plugins and be used
//	by all to communicate with their corresponding server-side logic.
//
package client

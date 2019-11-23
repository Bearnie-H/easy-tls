// Package client implements a minor extension of the default http.Client.
//
// The primary extensions of this implementation over the standard library are:
//
//	1) Integration with the base EasyTLS package, to provide a standardized and simpler interface for configuring and utilizing TLS encryption on an HTTP Client.
//
//	2) Explicit exposure and availability of all standard HTTP methods with a common interface.
//	    This is helpful when designing RESTful interfaces, where it can be more logically oriented to explicitly call a "Delete" function, rather than needing to create a specific DELETE request and submitting that to a generic "Do" function.  This is entirely an interface-type extension, as this package does include a generic "Do".
//	    This also allows a small amount of memory safety to be added.  There are a number of HTTP methods which do not expect a full HTTP Response with a meaningful body.  The standard library expects you to close these response bodies in all cases, while this library will handle closing response bodies for the cases where they are not meaningful.  This can help remove a memory leak issue which isn't immediately obvious.
//
//	3) Plugin integration via the easy-tls/plugins package.  Client/Server application design can be simplified or broken down from monoliths with an easy enough plugin system, which this is intended to provide.  The underlying http.Client is safe for concurrent use, therefore the plugin architecture assumes that a single SimpleClient instance can be passed in to any number of plugins and be used by all to communicate with their corresponding server-side logic.
//
package client

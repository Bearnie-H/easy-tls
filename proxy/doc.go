// Package proxy implements the extensions to a SimpleServer necessary
// to implement a full HTTP(S) Reverse Proxy.
//
// This package relies heavily on the "server" and "client" packages of this
// library to perform the actual routing and Request/Response Handling.
// The primary additional logic of this package is the "doReverseProxy"
// function, making use of a "ReverseProxyRouterFunc".
//
// The high-level idea of how this package operates is that all incoming
// requests which the proxy server is configured to accept (URI's match
// the path prefix) will be processed by the given "ReverseProxyRouterFunc".
// This function must be able to look at the current request and determine the
// host to forward to from it. This function MUST return the new host to
// forward to. The proxy server then uses a SimpleClient to perform the request
// before writing the response back to the original request source.
//
package proxy

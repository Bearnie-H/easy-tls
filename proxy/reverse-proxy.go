package proxy

import (
	"fmt"
	"html"
	"io"
	"log"
	"net/http"

	"github.com/Bearnie-H/easy-tls/client"
	"github.com/Bearnie-H/easy-tls/header"
	"github.com/Bearnie-H/easy-tls/server"
)

// NotFoundHandlerProxyOverride will override the NotFound handler of the
// Server with a reverse proxy lookup function. This will allow the server
// to attempt to re-route requests it doesn't have a defined route for, while
// still falling back to a "NotFound" 404 response if there is
// no defined place to route to.
func NotFoundHandlerProxyOverride(S *server.SimpleServer, c *client.SimpleClient, RouteMatcher ReverseProxyRouterFunc, logger *log.Logger) {

	var err error

	if logger == nil {
		logger = S.Logger()
	}

	if c == nil {
		c, err = client.NewClientHTTPS(S.TLSBundle())
		if err != nil {
			panic(err)
		}
		c.SetLogger(logger)
	}

	S.Router().NotFoundHandler = DoReverseProxy(c, RouteMatcher, logger)
}

// ConfigureReverseProxy will convert a freshly created SimpleServer
// into a ReverseProxy. This will use the provided SimpleClient
// (or a default HTTP SimpleClient) to perform the requests.
// The ReverseProxyRouterFunc defines HOW the routing will be performed, and
// must map individual requests to URLs to forward to.
// The PathPrefix defines the base path to proxy from, with a default of "/"
// indicating that ALL incoming requests should be proxied.
// If No Server or Client are provided, default instances will be generated.
func ConfigureReverseProxy(S *server.SimpleServer, Client *client.SimpleClient, logger *log.Logger, RouteMatcher ReverseProxyRouterFunc, PathPrefix string) *server.SimpleServer {

	// If No server is provided, create a default HTTP Server.
	var err error
	if S == nil {
		S = server.NewServerHTTP()
	}

	// Assert a non-empty path prefix to proxy on
	if PathPrefix == "" {
		PathPrefix = "/"
	}

	// If there's no logger provided, use the one from the server
	if logger == nil {
		logger = S.Logger()
	}

	// If no client is given, attempt to create one, using any TLS resources the potential server had.
	if Client == nil {
		Client, err = client.NewClientHTTPS(S.TLSBundle())
		if err != nil {
			panic(err)
		}
		Client.SetLogger(logger)
	}

	S.AddSubrouter(
		S.Router(),
		PathPrefix,
		server.NewSimpleHandler(
			DoReverseProxy(
				Client,
				RouteMatcher,
				logger,
			),
			PathPrefix,
		),
	)

	return S
}

// DoReverseProxy is the backbone of this package, and the reverse
// Proxy behaviour in general.
//
// This is the http.HandlerFunc which is called on ALL incoming requests
//  to the reverse proxy. At a high level this function:
//
//	1) Determines the forward host, from the incoming request
//	2) Creates a NEW request, performing a deep copy of the original, including the body
//	3) Performs this new request, using the provided (or default) SimpleClient to the new Host.
//	4) Receives the corresponding response, and deep copies it back to the original requester.
//
func DoReverseProxy(C *client.SimpleClient, Matcher ReverseProxyRouterFunc, logger *log.Logger) http.HandlerFunc {

	// Anonymous function to be returned, and is what is actually called when requests come in.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		// Create the new URL to use, based on the TLS settings of the Client, and the incoming request.
		proxyURL, err := Matcher(r)
		switch err {
		case nil:
		case ErrRouteNotFound:
			logger.Printf("Failed to find destination host:port for URL [ %s ] from %s - %s", r.URL.String(), r.RemoteAddr, err)
			w.WriteHeader(http.StatusNotFound)
			return
		case ErrForbiddenRoute:
			logger.Printf("Cannot forward request for URL [ %s ] from %s - %s", r.URL.String(), r.RemoteAddr, err)
			w.WriteHeader(http.StatusForbidden)
			return
		default:
			logger.Printf("Failed to format proxy forwarding for URL [ %s ] from %s - %s", r.URL.String(), r.RemoteAddr, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Create the new Request to send
		proxyReq, err := client.NewRequest(r.Method, proxyURL.String(), r.Header, r.Body)
		if err != nil {
			logger.Printf("Failed to create proxy forwarding request for URL [ %s ] from %s - %s", r.URL.String(), r.RemoteAddr, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Add in some proxy-specific headers
		proxyHeaders := http.Header{
			"Host":            []string{r.Host},
			"X-Forwarded-For": []string{r.RemoteAddr},
		}
		header.Merge(&(proxyReq.Header), &proxyHeaders)

		logger.Printf("Forwarding [ %s [ %s ] ] from [ %s ] to [ %s ]", r.URL.String(), r.Method, r.RemoteAddr, proxyURL.String())

		// Perform the full proxy request
		proxyResp, err := C.Do(proxyReq)
		if err != nil {
			logger.Printf("Failed to perform proxy request for URL [ %s ] from %s - %s", r.URL.String(), r.RemoteAddr, err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(html.EscapeString(fmt.Sprintf("Failed to perform proxy request for URL [ %s ] - %s.\n", r.URL.String(), err))))
			return
		}

		// Write the response fields out to the original requester
		responseHeader := w.Header()
		header.Merge(&responseHeader, &(proxyResp.Header))

		// Write back the status code
		w.WriteHeader(proxyResp.StatusCode)

		// Write back the response body
		if _, err := io.Copy(w, proxyResp.Body); err != nil {
			logger.Printf("Failed to write back proxy response for URL [ %s ] from %s - %s", r.URL.String(), r.RemoteAddr, err)
			return
		}
	})
}

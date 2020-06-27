package proxy

import (
	"fmt"
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

	if c == nil {
		c, err = client.NewClientHTTPS(S.TLSBundle(), client.DowngradeWithReset)
		if err != nil {
			panic(err)
		}
	}

	S.Router().NotFoundHandler = DoReverseProxy(c, RouteMatcher, logger)
}

// ConfigureReverseProxy will convert a freshly created SimpleServer
// into a ReverseProxy. This will use the provided SimpleClient
// (or a default HTTP SimpleClient) to perform the requests.
// The ReverseProxyRouterFunc defines HOW the routing will be peformed, and
// must map individual requests to URLs to forward to.
// The PathPrefix defines the base path to proxy from, with a default of "/"
// indicating that ALL incoming requests should be proxied.
// Finally, any middlewares desired can be added, noting that the
// "MiddlewareDefaultLogger" is applied in all cases.
// If No Server or Client are provided, default instances will be generated.
func ConfigureReverseProxy(S *server.SimpleServer, Client *client.SimpleClient, logger *log.Logger, RouteMatcher ReverseProxyRouterFunc, PathPrefix string, Middlewares ...server.MiddlewareHandler) {

	// If No server is provided, create a default HTTP Server.
	var err error
	if S == nil {
		S, err = server.NewServerHTTP()
		if err != nil {
			panic(err)
		}
	}

	// Assert a non-empty path prefix to proxy on
	if PathPrefix == "" {
		PathPrefix = "/"
	}

	// If no client is given, attempt to create one, using any TLS resources the potential server had.
	if Client == nil {
		Client, err = client.NewClientHTTPS(S.TLSBundle(), client.DowngradeWithReset)
		if err != nil {
			panic(err)
		}
	}

	S.Router().PathPrefix(PathPrefix).HandlerFunc(DoReverseProxy(Client, RouteMatcher, logger))
	S.AddMiddlewares(server.MiddlewareDefaultLogger(logger))
	S.AddMiddlewares(Middlewares...)
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
			if logger != nil {
				logger.Printf("Failed to find destination host:port for URL [ %s ] from %s - %s", r.URL.String(), r.RemoteAddr, err)
			}
			w.WriteHeader(http.StatusNotFound)
			return
		case ErrForbiddenRoute:
			if logger != nil {
				logger.Printf("Cannot forward request for URL [ %s ] from %s - %s", r.URL.String(), r.RemoteAddr, err)
			}
			w.WriteHeader(http.StatusForbidden)
			return
		default:
			if logger != nil {
				logger.Printf("Failed to format proxy forwarding for URL [ %s ] from %s - %s", r.URL.String(), r.RemoteAddr, err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Create the new Request to send
		proxyReq, err := client.NewRequest(r.Method, proxyURL, r.Header, r.Body)
		if err != nil {
			if logger != nil {
				logger.Printf("Failed to create proxy forwarding request for URL [ %s ] from %s - %s", r.URL.String(), r.RemoteAddr, err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Add in some proxy-specific headers
		proxyHeaders := http.Header{
			"Host":            []string{r.Host},
			"X-Forwarded-For": []string{r.RemoteAddr},
		}
		header.Merge(&(proxyReq.Header), &proxyHeaders)

		if logger != nil {
			logger.Printf("Forwarding [ %s [ %s ] ] from [ %s ] to [ %s ]", r.URL.String(), r.Method, r.RemoteAddr, proxyURL.String())
		}

		// Perform the full proxy request
		proxyResp, err := C.Do(proxyReq)
		switch err {
		case nil:
			defer proxyResp.Body.Close()
			break
		case client.ErrInvalidStatusCode:
			defer proxyResp.Body.Close()
			if logger != nil {
				logger.Printf("Failed to perform proxy request for URL [ %s ] from %s - %s", r.URL.String(), r.RemoteAddr, err)
			}
			proxyResp.Header.Del("Content-Length")
			H := w.Header()
			header.Merge(&H, &proxyResp.Header)
			w.WriteHeader(proxyResp.StatusCode)
			w.Write([]byte(fmt.Sprintf("Failed to perform proxy request for URL [ %s ] - %s.\n", r.URL.String(), err)))
			if _, err := io.Copy(w, proxyResp.Body); err != nil {
				if logger != nil {
					logger.Printf("Failed to write back failed proxy response for URL [ %s ] from %s - %s", r.URL.String(), r.RemoteAddr, err)
				}
				return
			}
			return
		default:
			if logger != nil {
				logger.Printf("Failed to perform proxy request for URL [ %s ] from %s - %s", r.URL.String(), r.RemoteAddr, err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Failed to perform proxy request for URL [ %s ] - %s.\n", r.URL.String(), err)))
			return
		}

		// Write the response fields out to the original requester
		responseHeader := w.Header()
		header.Merge(&responseHeader, &(proxyResp.Header))

		// Write back the status code
		w.WriteHeader(proxyResp.StatusCode)

		// Write back the response body
		if _, err := io.Copy(w, proxyResp.Body); err != nil {
			if logger != nil {
				logger.Printf("Failed to write back proxy response for URL [ %s ] from %s - %s", r.URL.String(), r.RemoteAddr, err)
			}
			return
		}
	})
}

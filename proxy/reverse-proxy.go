package proxy

import (
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/Bearnie-H/easy-tls/client"
	"github.com/Bearnie-H/easy-tls/server"
	"github.com/gorilla/mux"
)

// NotFoundHandlerProxyOverride will override the NotFound handler of the Server with a reverse proxy lookup function.
// This will allow the server to attempt to re-route requests it doesn't have a defined route for, while still falling back to a "NotFound" 404 response if there is no defined place to route to.
func NotFoundHandlerProxyOverride(r *mux.Router, RouteMatcher ReverseProxyRouterFunc, Verbose bool) {
	r.NotFoundHandler = DoReverseProxy(nil, false, RouteMatcher, Verbose)
}

// ConfigureReverseProxy will convert a freshly created SimpleServer into a ReverseProxy.  This will use either the provided SimpleClient (or a default HTTP SimpleClient) to perform the requests.  The ReverseProxyRouterFunc defines HOW the routing will be peformed, and must map individual requests to URLs to forward to.  The PathPrefix defines the base path to proxy from, with a default of "/" indicating that ALL incoming requests should be proxied.  Finally, any middlewares desired can be added, noting that the "MiddlewareDefaultLogger" is applied in all cases.
//
// If No Server or Client are provided, default instances will be generated.
func ConfigureReverseProxy(S *server.SimpleServer, Client *client.SimpleClient, Verbose bool, RouteMatcher ReverseProxyRouterFunc, PathPrefix string, Middlewares ...server.MiddlewareHandler) {

	// If No server is provided, create a default HTTP Server.
	var err error
	if S == nil {
		S, err = server.NewServerHTTP()
		if err != nil {
			panic(err)
		}
	}

	if PathPrefix == "" {
		PathPrefix = "/"
	}

	r := server.NewDefaultRouter()

	if Client == nil {
		Client, err = client.NewClientHTTP()
		if err != nil {
			panic(err)
		}
	}

	r.PathPrefix(PathPrefix).HandlerFunc(DoReverseProxy(Client, Client.IsTLS(), RouteMatcher, Verbose))

	server.AddMiddlewares(r, server.MiddlewareDefaultLogger)
	server.AddMiddlewares(r, Middlewares...)

	S.RegisterRouter(r)
}

// DoReverseProxy is the backbone of this package, and the reverse Proxy behaviour in general.
//
// This is the http.HandlerFunc which is called on ALL incoming requests to the reverse proxy.  At a high level this function:
//	1) Determines the forward host, from the incoming request
//	2) Creates a NEW request, performing a deep copy of the original, excluding the body
//	3) Performs this new request, using the provided (or default) SimpleClient to the new Host.
//	4) Receives the corresponding response, and deep copies it back to the original requester.
func DoReverseProxy(C *client.SimpleClient, IsTLS bool, Matcher ReverseProxyRouterFunc, verbose bool) http.HandlerFunc {

	// If no client is provided, create a default HTTP Client to perform the requests.
	if C == nil {
		var err error
		C, err = client.NewClientHTTP()
		if err != nil {
			panic(err)
		}
	}

	// Anonymous function to be returned, and is what is actually called when requests come in.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Create the new URL to use, based on the TLS settings of the Client, and the incoming request.
		proxyURL, err := formatProxyURL(r, IsTLS, Matcher)
		if err == ErrRouteNotFound {
			log.Printf("Failed to find destination host:port for URL %s from %s - %s", r.URL.EscapedPath(), r.RemoteAddr, err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err == ErrForbiddenRoute {
			log.Printf("Cannot forward route %s - %s", r.URL.EscapedPath(), err)
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if err != nil {
			log.Printf("Failed to format proxy forwarding URL from %s - %s", r.RemoteAddr, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Create the new Request to send
		proxyReq, err := client.NewRequest(r.Method, proxyURL, r.Header, r.Body)
		if err != nil {
			log.Printf("Failed to create proxy forwarding request for %s from %s - %s", r.URL.String(), r.RemoteAddr, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Add in some proxy-specific headers
		proxyReq.Header.Add("Host", r.Host)
		proxyReq.Header.Add("X-Forwarded-For", r.RemoteAddr)

		if verbose {
			log.Printf("Forwarding %s to %s%s", r.URL.Path, proxyURL.Host, proxyURL.Path)
		}

		// Perform the full proxy request
		proxyResp, err := C.Do(proxyReq)
		if err != nil {
			log.Printf("Failed to perform proxy request for %s from %s - %s", r.URL.String(), r.RemoteAddr, err)
			if proxyResp != nil {
				w.WriteHeader(proxyResp.StatusCode)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
		defer proxyResp.Body.Close()

		// Write the response fields out to the original requester
		for header, values := range proxyResp.Header {
			for _, value := range values {
				w.Header().Add(header, value)
			}
		}

		// Write back the status code
		w.WriteHeader(proxyResp.StatusCode)

		// Write back the response body
		if _, err := io.Copy(w, proxyResp.Body); err != nil {
			log.Printf("Failed to write back proxy response for %s from %s - %s", r.URL.String(), r.RemoteAddr, err)
			return
		}
	})
}

// formatProxyURL will look at the original request, the TLS Settings of the SimpleClient, and generate what the new URL must be for the proxied request.
func formatProxyURL(req *http.Request, IsTLS bool, MatcherFunc ReverseProxyRouterFunc) (*url.URL, error) {
	var err error
	proxyURL := &url.URL{}

	// Deep Copy
	*proxyURL = *req.URL

	proxyURL.Host, proxyURL.Path, err = MatcherFunc(req)
	if err != nil {
		return nil, err
	}

	proxyURL.Scheme = "http"
	if IsTLS {
		proxyURL.Scheme = "https"
	}

	return proxyURL, nil
}

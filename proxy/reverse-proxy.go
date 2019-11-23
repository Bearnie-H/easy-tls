package proxy

import (
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/Bearnie-H/easy-tls/client"
	"github.com/Bearnie-H/easy-tls/server"
)

// ConfigureReverseProxy will convert a freshly created SimpleServer into a ReverseProxy.  This will use either the provided SimpleClient (or a default HTTP SimpleClient) to perform the requests.  The ReverseProxyRouterFunc defines HOW the routing will be peformed, and must map individual requests to URLs to forward to.  The PathPrefix defines the base path to proxy from, with a default of "/" indicating that ALL incoming requests should be proxied.  Finally, any middlewares desired can be added, noting that the "MiddlewareDefaultLogger" is applied in all cases.
func ConfigureReverseProxy(S *server.SimpleServer, Client *client.SimpleClient, RouteMatcher ReverseProxyRouterFunc, PathPrefix string, Middlewares ...server.MiddlewareHandler) {

	r := server.NewDefaultRouter()

	r.PathPrefix(PathPrefix).HandlerFunc(doReverseProxy(Client, Client.IsTLS(), RouteMatcher))

	server.AddMiddlewares(r, server.MiddlewareDefaultLogger)
	server.AddMiddlewares(r, Middlewares...)

	S.RegisterRouter(r)
}

// ReverseProxyRouterFunc represents the Type which must be satisfied by any function which defines the per-request routing behaviours.  This must map a given request to a specific IP:Port host.
type ReverseProxyRouterFunc func(*http.Request) string

// doReverseProxy is the backbone of this package, and the reverse Proxy behaviour in general.
//
// This is the http.HandlerFunc which is called on ALL incoming requests to the reverse proxy.  At a high level this function:
//	1) Determines the forward host, from the incoming request
//	2) Creates a NEW request, performing a deep copy of the original, excluding the body
//	3) Performs this new request, using the provided (or default) SimpleClient to the new Host.
//	4) Receives the corresponding response, and deep copies it back to the original requester.
func doReverseProxy(C *client.SimpleClient, IsTLS bool, Matcher ReverseProxyRouterFunc) http.HandlerFunc {

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
		proxyURL := formatProxyURL(r, IsTLS, Matcher)

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

		// Perform the full proxy request
		proxyResp, err := C.Do(proxyReq)
		if err != nil {
			log.Printf("Failed to perform proxy request for %s from %s - %s", r.URL.String(), r.RemoteAddr, err)
			w.WriteHeader(http.StatusInternalServerError)
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
func formatProxyURL(req *http.Request, IsTLS bool, MatcherFunc ReverseProxyRouterFunc) *url.URL {
	proxyURL := &url.URL{}

	// Deep Copy
	*proxyURL = *req.URL

	proxyURL.Host = MatcherFunc(req)
	proxyURL.Scheme = "http"
	if IsTLS {
		proxyURL.Scheme = "https"
	}

	return proxyURL
}

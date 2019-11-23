package proxy

import (
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/Bearnie-H/easy-tls/client"
	"github.com/Bearnie-H/easy-tls/server"
)

// ConfigureReverseProxy will convert a freshly created SimpleServer into a ReverseProxy, forwarding all incoming traffic based on the RouteMatcher func provided.  This will create the necessary HTTP handler, and configure the necessary routing.
func ConfigureReverseProxy(S *server.SimpleServer, Client *client.SimpleClient, RouteMatcher ReverseProxyRouterFunc, PathPrefix string, Middlewares ...server.MiddlewareHandler) {

	r := server.NewDefaultRouter()

	r.PathPrefix(PathPrefix).HandlerFunc(doReverseProxy(Client, Client.IsTLS(), RouteMatcher))

	server.AddMiddlewares(r, server.MiddlewareDefaultLogger)
	server.AddMiddlewares(r, Middlewares...)

	S.RegisterRouter(r)
}

// ReverseProxyRouterFunc will take a request, and determine which URL Host to forward it to.  This result must be an IP:Port combination as standard in the http package.
type ReverseProxyRouterFunc func(*http.Request) string

// doReverseProxy will forward all traffic coming in through the SimpleClient, swapping to/from TLS as specified by the SimpleClient, and determining which remote host to forward to based on the Matcher function.  This provides an opaque connection, and neither side should know they are talking through a proxy, aside from the headers explicitly placed into the ProxyRequest.
func doReverseProxy(C *client.SimpleClient, IsTLS bool, Matcher ReverseProxyRouterFunc) http.HandlerFunc {

	// If no client is provided, create one.
	if C == nil {
		var err error
		C, err = client.NewClientHTTP()
		if err != nil {
			panic(err)
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Create the new URL to use
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

		// Write back the header fields
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

func formatProxyURL(req *http.Request, IsTLS bool, MatcherFunc ReverseProxyRouterFunc) *url.URL {
	proxyURL := &url.URL{}
	*proxyURL = *req.URL
	proxyURL.Host = MatcherFunc(req)
	if IsTLS {
		proxyURL.Scheme = "https"
	} else {
		proxyURL.Scheme = "http"
	}

	return proxyURL
}

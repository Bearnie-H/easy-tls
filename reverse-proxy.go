package easytls

import (
	"log"
	"net/http"
)

// ReverseProxyRouterFunc will take a request, and determine which URL Host to forward it to.  This result must be an IP:Port combination as standard in the http package.
type ReverseProxyRouterFunc func(*http.Request) string

func doReverseProxy(C *SimpleClient, IsTLS bool, Matcher ReverseProxyRouterFunc) http.HandlerFunc {

	// If no client is provided, create one.
	if C == nil {
		var err error
		C, err = NewClientHTTP()
		if err != nil {
			panic(err)
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		IncomingURL := *r.URL
		url := r.URL
		url.Host = Matcher(r)
		if IsTLS {
			url.Scheme = "https"
		} else {
			url.Scheme = "http"
		}

		log.Printf("Got incoming request with URL: %s, forwarding to: %s", IncomingURL.String(), url.String())

		proxyReq, err := http.NewRequest(r.Method, url.String(), r.Body)
		if err != nil {
			log.Printf("Failed to create proxy request - %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		proxyReq.Header.Set("Host", r.Host)
		proxyReq.Header.Set("X-Forwarded-For", r.RemoteAddr)

		for header, values := range r.Header {
			for _, value := range values {
				proxyReq.Header.Add(header, value)
			}
		}

		proxyResp, err := C.Do(proxyReq)
		if err != nil {
			log.Printf("Failed to perform proxy request - %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := proxyResp.Write(w); err != nil {
			log.Printf("Failed to write back proxy response - %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}

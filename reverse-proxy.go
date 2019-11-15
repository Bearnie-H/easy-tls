package easytls

import (
	"log"
	"net/http"
	"net/url"
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
		var url *url.URL
		*url = *r.URL
		url.Host = Matcher(r)
		if IsTLS {
			url.Scheme = "https"
		} else {
			url.Scheme = "http"
		}

		proxyReq, err := http.NewRequest(r.Method, url.String(), r.Body)
		if err != nil {
			log.Printf("Failed to create proxy forwarding request for %s from %s - %s", r.URL.String(), r.Host, err)
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
			log.Printf("Failed to perform proxy request for %s from %s - %s", r.URL.String(), r.Host, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := proxyResp.Write(w); err != nil {
			log.Printf("Failed to write back proxy response for %s from %s - %s", r.URL.String(), r.Host, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}

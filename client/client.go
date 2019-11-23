package client

import (
	"io"
	"net/http"
	"net/url"
	"time"

	easytls "github.com/Bearnie-H/easy-tls"
)

// SimpleClient is an extension of the standard http.Client implementation, with additional utility functions and wrappers to simplify using it.
type SimpleClient struct {
	client *http.Client
	tls    bool
}

// NewClientHTTP will create a new SimpleClient, with no TLS settings enabled.  This will accept raw HTTP only.
func NewClientHTTP() (*SimpleClient, error) {
	return NewClientHTTPS(nil)
}

// NewClientHTTPS will create a new TLS-Enabled SimpleClient.  This will
func NewClientHTTPS(TLS *easytls.TLSBundle) (*SimpleClient, error) {
	tls, err := easytls.NewTLSConfig(TLS)
	if err != nil {
		return nil, err
	}

	s := &SimpleClient{
		client: &http.Client{
			Timeout: time.Hour,
			Transport: &http.Transport{
				TLSClientConfig:   tls,
				ForceAttemptHTTP2: true,
			}},
		tls: !(TLS == nil),
	}

	return s, nil
}

// IsTLS exposes whether the SimpleClient is TLS or not.
func (C *SimpleClient) IsTLS() bool {
	return C.tls
}

// NewRequest will create a new HTTP Request, ready to be used by any implementation of an http.Client
func NewRequest(Method string, URL *url.URL, Headers map[string][]string, Contents io.ReadCloser) (*http.Request, error) {
	req, err := http.NewRequest(Method, URL.String(), Contents)
	if err != nil {
		return nil, err
	}

	for k, vs := range Headers {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}

	return req, nil
}

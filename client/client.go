package client

import (
	"crypto/tls"
	"io"
	"net/http"
	"net/url"
	"time"

	easytls "github.com/Bearnie-H/easy-tls"
	"github.com/Bearnie-H/easy-tls/common"
)

// SimpleClient is the primary object of this library.  This is the implementation of the simplified HTTP Client provided by this package.  The use and functionality of this is opaque to whether or not this is running in HTTP or HTTPS mode, with a basic utility function to check.
type SimpleClient struct {
	client *http.Client
	tls    bool
}

// NewClientHTTP will fully initialize a SimpleClient with TLS settings turned off.  These settings CAN be turned on and off as required.
func NewClientHTTP() (*SimpleClient, error) {
	return NewClientHTTPS(nil)
}

// NewClientHTTPS will fully initialize a SimpleClient with TLS settings turned on.  These settings CAN be turned on and off as required.
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

// IsTLS returne whether the SimpleClient is currently TLS-enabled or not.
func (C *SimpleClient) IsTLS() bool {
	return C.tls
}

// NewRequest will create a new HTTP Request, ready to be used by any implementation of an http.Client
func NewRequest(Method string, URL *url.URL, Headers map[string][]string, Contents io.ReadCloser) (*http.Request, error) {
	req, err := http.NewRequest(Method, URL.String(), Contents)
	if err != nil {
		return nil, err
	}

	common.AddHeaders(&(req.Header), Headers)

	return req, nil
}

// EnableTLS will enable the TLS settings for a SimpleClient based on the provided TLSBundle.
func (C *SimpleClient) EnableTLS(TLS *easytls.TLSBundle) error {
	tls, err := easytls.NewTLSConfig(TLS)
	if err != nil {
		return err
	}

	C.client = &http.Client{
		Timeout: time.Hour,
		Transport: &http.Transport{
			TLSClientConfig:   tls,
			ForceAttemptHTTP2: true,
		},
	}
	C.tls = true

	return nil
}

// DisableTLS will turn off the TLS settings for a SimpleClient.
func (C *SimpleClient) DisableTLS() {
	C.client = &http.Client{
		Timeout: time.Hour,
		Transport: &http.Transport{
			TLSClientConfig:   &tls.Config{},
			ForceAttemptHTTP2: true,
		},
	}
	C.tls = false
}

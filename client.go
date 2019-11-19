package easytls

import (
	"net/http"
	"time"
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
func NewClientHTTPS(TLS *TLSBundle) (*SimpleClient, error) {
	tls, err := NewTLSConfig(TLS)
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

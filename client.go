package easytls

import (
	"net/http"
	"time"
)

// SimpleClient is a renaming of the Standard http.Client for this package, to allow the ease-of-use extensions provided here.
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

	Enabled := !(TLS == nil)

	s := &SimpleClient{
		client: &http.Client{
			Timeout:   time.Hour * 1,
			Transport: &http.Transport{TLSClientConfig: tls}},
		tls: Enabled,
	}

	return s, nil
}

// IsTLS exposes whether the SimpleClient is TLS or not.
func (C *SimpleClient) IsTLS() bool {
	return C.tls
}

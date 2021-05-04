package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	easytls "github.com/Bearnie-H/easy-tls"
	"github.com/Bearnie-H/easy-tls/header"
)

// SimpleClient is the primary object of this library. This is the
// implementation of the simplified HTTP Client provided by this package.
// The use and functionality of this is transparent to whether or not this is
// running in HTTP or HTTPS mode, with a basic utility function to check.
type SimpleClient struct {
	*http.Client
	logger *log.Logger

	tls    bool
	bundle easytls.TLSBundle
}

// NewClient will wrap an existing http.Client as a SimpleClient.
func NewClient(C *http.Client) *SimpleClient {
	return &SimpleClient{
		Client: C,
		logger: easytls.NewDefaultLogger(),
		tls:    false,
		bundle: easytls.TLSBundle{},
	}
}

// NewClientHTTP will fully initialize a SimpleClient with TLS settings turned
// off. These settings CAN be turned on and off as required, either by
// providing a TLSBundle, or by reusing one passed in earlier.
func NewClientHTTP() *SimpleClient {
	C, _ := NewClientHTTPS(nil)
	return C
}

// NewClientHTTPS will fully initialize a SimpleClient with TLS settings turned
// on. These settings CAN be turned on and off as required.
func NewClientHTTPS(TLS *easytls.TLSBundle) (*SimpleClient, error) {

	tls, err := easytls.NewTLSConfig(TLS)
	if err != nil {
		return nil, err
	}

	Logger := easytls.NewDefaultLogger()

	var saveBundle easytls.TLSBundle
	if TLS != nil {
		saveBundle = *TLS
	}

	s := &SimpleClient{
		Client: &http.Client{
			Timeout: time.Hour,
			Transport: &http.Transport{
				TLSClientConfig:   tls,
				ForceAttemptHTTP2: true,
			}},
		tls:    !(tls == nil),
		logger: Logger,
		bundle: saveBundle,
	}

	return s, nil
}

// SetLogger will update the logger used by the client from the default to the
// given output.
func (C *SimpleClient) SetLogger(logger *log.Logger) {
	C.logger = logger
}

// Logger will return the internal logger used by the client.
func (C *SimpleClient) Logger() *log.Logger {
	return C.logger
}

// CloneTLSConfig will form a proper clone of the underlying tls.Config.
func (C *SimpleClient) CloneTLSConfig() (*tls.Config, error) {
	return easytls.NewTLSConfig(&C.bundle)
}

// IsTLS returns whether the SimpleClient is currently TLS-enabled or not.
func (C *SimpleClient) IsTLS() bool {
	return C.tls
}

// MakeURL will create a HTTP URL, ready to be used to build a Request.
func (C *SimpleClient) MakeURL(Hostname string, Port uint16, PathSegments ...string) *url.URL {

	scheme := "http"
	if C.IsTLS() {
		scheme = "https"
	}

	return &url.URL{
		Scheme: scheme,
		Host:   fmt.Sprintf("%s:%d", Hostname, Port),
		Path:   strings.Join(PathSegments, "/"),
	}
}

// NewRequest will create a new HTTP Request, ready to be used by any
// implementation of an http.Client.
func NewRequest(Method string, URL string, Headers http.Header, Contents io.Reader) (*http.Request, error) {
	return NewRequestWithContext(context.Background(), Method, URL, Headers, Contents)
}

// NewRequestWithContext will create a new HTTP Request, ready to be used by any
// implementation of an http.Client.
func NewRequestWithContext(ctx context.Context, Method string, URL string, Headers http.Header, Contents io.Reader) (*http.Request, error) {

	req, err := http.NewRequestWithContext(ctx, Method, URL, Contents)
	if err != nil {
		return nil, err
	}

	header.Merge(&(req.Header), &Headers)

	return req, nil
}

// EnableTLS will enable the TLS settings for a SimpleClient based on the
// provided TLSBundle. If the client previously had a TLS bundle provided,
// this will fall back and attempt to use that if none is given. If no
// TLSBundles are given, and the Client has no previous TLSBundle, this will
// fail, as there are no TLS resources to work with.
func (C *SimpleClient) EnableTLS(TLS ...*easytls.TLSBundle) (err error) {

	var tlsConf *tls.Config
	for _, tls := range TLS {
		tlsConf, err = easytls.NewTLSConfig(tls)
		if err != nil {
			continue
		}
	}

	// If all of the provided TLSBundles failed, check if the Client has a saved one and use it.
	if err != nil {
		tlsConf, err = easytls.NewTLSConfig(&C.bundle)
		if err != nil {
			return err
		}
	}

	C.Client = &http.Client{
		Timeout: time.Hour,
		Transport: &http.Transport{
			TLSClientConfig:   tlsConf,
			ForceAttemptHTTP2: true,
		},
	}
	C.tls = true

	return nil
}

// DisableTLS will turn off the TLS settings for a SimpleClient.
func (C *SimpleClient) DisableTLS() {

	C.Client = &http.Client{
		Timeout: time.Hour,
		Transport: &http.Transport{
			TLSClientConfig:   &tls.Config{},
			ForceAttemptHTTP2: true,
		},
	}
	C.tls = false
}

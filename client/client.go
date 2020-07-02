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

// TLSRetryPolicy defines the possible policies that a SimpleClient can take
// when it communicates with the wrong HTTP/HTTPS to the server.
type TLSRetryPolicy int

// Enum definition of the available TLSRetryPolicy values
// NOTE: Currently only NoRetry and Downgrade are implemented.
const (

	// Don't retry the attempt using a different Scheme
	NoRetry TLSRetryPolicy = iota

	// Attempt only downgrading from HTTPS to HTTP, and don't reset the
	// SimpleClient after the request.
	DowngradeNoReset

	// Attempt only downgrading from HTTPS to HTTP, and reset the SimpleClient
	// after the request.
	DowngradeWithReset

	// Attempt only upgrading from HTTP to HTTPS, but reset the SimpleClient to
	// HTTP after the request.
	UpgradeNoReset

	// Attempt only upgrading from HTTP to HTTPS, and keep the SimpleClient
	// configured for HTTPS after the request.
	UpgradeWithReset

	// Attempt to swap HTTP <-> HTTPS, and don't reset the SimpleClient after.
	SwapNoReset

	// Attempt to swap HTTP <-> HTTPS, and do reset the SimpleClient after.
	SwapWithReset
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

	policy TLSRetryPolicy
}

// NewClient will wrap an existing http.Client as a SimpleClient.
func NewClient(C *http.Client) *SimpleClient {
	return &SimpleClient{
		Client: C,
		logger: easytls.NewDefaultLogger(),
		tls:    false,
		bundle: easytls.TLSBundle{},
		policy: NoRetry,
	}
}

// NewClientHTTP will fully initialize a SimpleClient with TLS settings turned
// off. These settings CAN be turned on and off as required, either by
// providing a TLSBundle, or by reusing one passed in earlier.
func NewClientHTTP() (*SimpleClient, error) {
	return NewClientHTTPS(nil, NoRetry)
}

// NewClientHTTPS will fully initialize a SimpleClient with TLS settings turned
// on. These settings CAN be turned on and off as required.
func NewClientHTTPS(TLS *easytls.TLSBundle, TLSPolicy TLSRetryPolicy) (*SimpleClient, error) {
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
		tls:    !(TLS == nil),
		logger: Logger,
		bundle: saveBundle,
		policy: TLSPolicy,
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
func NewRequest(Method string, URL *url.URL, Headers http.Header, Contents io.Reader) (*http.Request, error) {

	req, err := http.NewRequest(Method, URL.String(), Contents)
	if err != nil {
		return nil, err
	}

	header.Merge(&(req.Header), &Headers)

	return req, nil
}

// NewRequestWithContext will create a new HTTP Request, ready to be used by any
// implementation of an http.Client.
func NewRequestWithContext(ctx context.Context, Method string, URL *url.URL, Headers http.Header, Contents io.Reader) (*http.Request, error) {

	req, err := http.NewRequestWithContext(ctx, Method, URL.String(), Contents)
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

// retryWithDowngrade will attempt to re-send an HTTP request, after
// downgrading from HTTPS to HTTP. If this errors out, simply return the
// original response back, as if this never happpened.
//
// NOTE: This currently only works with requests that have empty bodies.
func (C *SimpleClient) retryWithDowngrade(r *http.Request, original *http.Response) (resp *http.Response, err error) {

	// Check the Upgrade/Downgrade policy
	switch C.policy {

	// If the policy implies resetting the TLS status, defer that to happen at the end
	case DowngradeWithReset, SwapWithReset:
		defer func(err *error) {
			err2 := C.EnableTLS()
			if *err != nil {
				if err2 != nil {
					*err = fmt.Errorf("%s. easytls client error: Failed to re-enable TLS - %s", *err, err2)
				}
			}
		}(&err) // Allow the reset of the TLS settings to report it's error back

	// If the policy doesn't specify resetting, don't defer resetting it
	case DowngradeNoReset, SwapNoReset:

		// If the policy is anything else, we shouldn't have gotten here, but just return and ignore this call.
	default:
		return original, nil
	}

	C.DisableTLS()

	newURL := *r.URL
	newURL.Scheme = "http"

	rewoundBody, err := C.rewindRequestBody(r)
	if err != nil {
		return original, fmt.Errorf("easytls client error: Failed to rewind request body in [ %s ] request to [ %s ] - %s", r.Method, r.URL.String(), err)
	}

	newReq, err := NewRequest(r.Method, &newURL, r.Header, rewoundBody)
	if err != nil {
		return original, fmt.Errorf("easytls client error: Failed to re-create [ %s ] request during TLS upgrade retry - %s", r.Method, err)
	}

	return C.Do(newReq)
}

// retryWithUpgrade will attempt to re-send an HTTP request, after upgrading
// from HTTP to HTTPS. If this errors out, simply return the original response
// back, as if this never happpened.
//
// NOTE: This doesn't work yet, as it's unclear how to generically identify
// when a client would need to attempt an upgrade.
func (C *SimpleClient) retryWithUpgrade(r *http.Request, original *http.Response) (*http.Response, error) {
	// Check the Upgrade/Downgrade policy
	switch C.policy {

	// If the policy implies resetting the TLS status, defer that to happen at the end
	case UpgradeWithReset, SwapWithReset:
		defer C.DisableTLS()

	// If the policy doesn't specify resetting, don't defer resetting it
	case UpgradeNoReset, SwapNoReset:

		// If the policy is anything else, we shouldn't have gotten here, but just return and ignore this call.
	default:
		return original, nil
	}

	if err := C.EnableTLS(); err != nil {
		return original, fmt.Errorf("easytls client error: Failed to enable TLS during retry attempt - %s", err)
	}

	newURL := *r.URL
	newURL.Scheme = "https"

	rewoundBody, err := C.rewindRequestBody(r)
	if err != nil {
		return original, fmt.Errorf("easytls client error: Failed to rewind request body in [ %s ] request to [ %s ] - %s", r.Method, r.URL.String(), err)
	}

	newReq, err := NewRequest(r.Method, &newURL, r.Header, rewoundBody)
	if err != nil {
		return original, fmt.Errorf("easytls client error: Failed to re-create [ %s ] request during TLS upgrade retry - %s", r.Method, err)
	}

	return C.Do(newReq)
}

func (C *SimpleClient) rewindRequestBody(r *http.Request) (io.ReadCloser, error) {

	// NOTE: This needs to be implemented. This should either attempt to rewind the request body if possible, or return a meaningful error.
	return r.Body, nil
}

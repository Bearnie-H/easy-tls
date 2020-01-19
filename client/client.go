package client

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	easytls "github.com/Bearnie-H/easy-tls"
	"github.com/Bearnie-H/easy-tls/header"
)

// TLSRetryPolicy defines the possible policies that a SimpleClient can take when it communicates with the wrong HTTP/HTTPS to the server
type TLSRetryPolicy int

// Enum definition of the available TLSRetryPolicy values
const (
	NoRetry            TLSRetryPolicy = iota // Don't retry the attempt using a different Scheme
	DowngradeNoReset                         // Attempt only downgrading from HTTPS to HTTP, and don't reset the SimpleClient after the request.
	DowngradeWithReset                       // Attempt only downgrading from HTTPS to HTTP, and reset the SimpleClient after the request.
	UpgradeNoReset                           // Attempt only upgrading from HTTP to HTTPS, but reset the SimpleClient to HTTP after the request.
	UpgradeWithReset                         // Attempt only upgrading from HTTP to HTTPS, and keep the SimpleClient configured for HTTPS after the request.
	SwapNoReset                              // Attempt to swap HTTP <-> HTTPS, and don't reset the SimpleClient after.
	SwapWithReset                            // Attempt to swap HTTP <-> HTTPS, and do reset the SimpleClient after.
)

// SimpleClient is the primary object of this library.  This is the implementation of the simplified HTTP Client provided by this package.  The use and functionality of this is opaque to whether or not this is running in HTTP or HTTPS mode, with a basic utility function to check.
type SimpleClient struct {
	client *http.Client

	tls    bool
	bundle easytls.TLSBundle

	policy TLSRetryPolicy
}

// NewClientHTTP will fully initialize a SimpleClient with TLS settings turned off.  These settings CAN be turned on and off as required.
func NewClientHTTP(TLSPolicy TLSRetryPolicy) (*SimpleClient, error) {
	if TLSPolicy != NoRetry {
		return nil, errors.New("easytls client error: Cannot initialize HTTP Client with TLS Retry Policy. Use NewClientHTTPS() or NoRetry instead")
	}
	return NewClientHTTPS(nil, TLSPolicy)
}

// NewClientHTTPS will fully initialize a SimpleClient with TLS settings turned on.  These settings CAN be turned on and off as required.
func NewClientHTTPS(TLS *easytls.TLSBundle, TLSPolicy TLSRetryPolicy) (*SimpleClient, error) {
	tls, err := easytls.NewTLSConfig(TLS)
	if err != nil {
		return nil, err
	}

	var saveBundle easytls.TLSBundle
	if TLS != nil {
		saveBundle = *TLS
	}

	s := &SimpleClient{
		client: &http.Client{
			Timeout: time.Hour,
			Transport: &http.Transport{
				TLSClientConfig:   tls,
				ForceAttemptHTTP2: true,
			}},
		tls:    !(TLS == nil),
		bundle: saveBundle,
		policy: TLSPolicy,
	}

	return s, nil
}

// IsTLS returne whether the SimpleClient is currently TLS-enabled or not.
func (C *SimpleClient) IsTLS() bool {
	return C.tls
}

// NewRequest will create a new HTTP Request, ready to be used by any implementation of an http.Client
func NewRequest(Method string, URL *url.URL, Headers http.Header, Contents io.Reader) (*http.Request, error) {

	req, err := http.NewRequest(Method, URL.String(), Contents)
	if err != nil {
		return nil, err
	}

	header.Merge(&(req.Header), &Headers)

	return req, nil
}

// EnableTLS will enable the TLS settings for a SimpleClient based on the provided TLSBundle.
// If the client previously had a TLS bundle provided, this will fall back and attempt to use it
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

	C.client = &http.Client{
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
	C.client = &http.Client{
		Timeout: time.Hour,
		Transport: &http.Transport{
			TLSClientConfig:   &tls.Config{},
			ForceAttemptHTTP2: true,
		},
	}
	C.tls = false
}

// retryWithDowngrade will attempt to re-send an HTTP request, after downgrading from HTTPS to HTTP.
// If this errors out, simply return the original response back, as if this never happpened.
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

// retryWithUpgrade will attempt to re-send an HTTP request, after upgrading from HTTP to HTTPS.
// If this errors out, simply return the original response back, as if this never happpened.
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

// This needs to be updated to work for non-empty request bodies.
func (C *SimpleClient) rewindRequestBody(r *http.Request) (io.ReadCloser, error) {
	// NOTE: This needs to be implemented.  This should either attempt to rewind the request body if possible, or return a meaningful error.
	return r.Body, nil
}

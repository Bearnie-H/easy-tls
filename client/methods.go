package client

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

// ErrInvalidStatusCode is the standard error to return when an HTTP request
// succeeds, but returns with a status code indicating some sort of error.
var ErrInvalidStatusCode = errors.New("Invalid status code - Expected 2xx")

// Set the URL Scheme based on the TLS settings of the Client.
func (C *SimpleClient) setScheme(URL *url.URL) {
	if C.IsTLS() {
		URL.Scheme = "https"
	} else {
		URL.Scheme = "http"
	}
}

// Get is the wrapper function for an HTTP "GET" request. This will create a
// GET request with an empty body and the specified headers. The header map can
// be set to nil if no additional headers are required. This function returns
// an error and nil response on an HTTP StatusCode which is outside the 200
// block.
func (C *SimpleClient) Get(URL *url.URL, Headers map[string][]string) (*http.Response, error) {

	// Set the scheme, allowing for this to be missing, and be able to be defined by the internal TLS state of the Client.
	C.setScheme(URL)

	// Create the request
	req, err := NewRequest(http.MethodGet, URL, Headers, nil)
	if err != nil {
		return nil, err
	}

	// Perform the request
	return C.Do(req)
}

// Head is the wrapper function for an HTTP "HEAD" request. This will create a
// new HEAD request with an empty body and the specified headers. The header
// map can be set to nil if no additional headers are required. This will ONLY
// return the HTTP Response Header map from the server. The overall Response
// Body (if it exists) will be closed by this function. This function returns
// an error and nil Header on an HTTP StatusCode which is outside the 200
// block.
func (C *SimpleClient) Head(URL *url.URL, Headers map[string][]string) (*http.Response, error) {

	// Set the scheme, allowing for this to be missing, and be able to be defined by the internal TLS state of the Client.
	C.setScheme(URL)

	// Create the request
	req, err := NewRequest(http.MethodHead, URL, Headers, nil)
	if err != nil {
		return nil, err
	}

	// Perform the request
	resp, err := C.Do(req)
	switch err {
	case nil:
		defer resp.Body.Close()
		return resp, nil
	case ErrInvalidStatusCode:
		defer resp.Body.Close()
		return resp, err
	default:
		return nil, err
	}
}

// Post is the wrapper function for an HTTP "POST" request. This will create a
// new POST request with a body composed of the contents of the io.Reader
// passed in, and the specified headers. The header map can be set to nil if no
// additional headers are required. If a nil ReadCloser is passed in, this will
// create an empty Post body which is allowed. This will return the full HTTP
// Response from the server, unaltered. This function returns an error and nil
// response on an HTTP StatusCode which is outside the 200 block.
//
// NOTE: This function "may" support MultiPart POST requests, by way of
// io.Pipes and multipart.Writers, but this has not been tested, and multipart
// Post is planned to be an explicit function in the future.
func (C *SimpleClient) Post(URL *url.URL, Contents io.Reader, Headers map[string][]string) (*http.Response, error) {

	// Set the scheme, allowing for this to be missing, and be able to be defined by the internal TLS state of the Client.
	C.setScheme(URL)

	// Create the request
	req, err := NewRequest(http.MethodPost, URL, Headers, Contents)
	if err != nil {
		return nil, err
	}

	// Perform the request
	return C.Do(req)
}

// PostMultipart is the wrapper function for an HTTP "POST" request with a
// MultiPart Body. This will create a new POST request with a body composed of
// the contents of the multipart.Reader passed in, and the specified headers.
// The header map can be set to nil if no additional headers are required. If
// a nil multipart.Reader is passed in, this will create an empty Post body
// which is allowed. This will return the full HTTP Response from the server
// unaltered. This function returns an error and nil response on an HTTP
// StatusCode which is outside the 200 block.
//
// NOTE: This has not yet been implemented.
func (C *SimpleClient) PostMultipart(URL *url.URL, Contents multipart.Reader, Headers map[string][]string) (*http.Response, error) {
	return nil, errors.New("Method POST-MULTIPART not yet implemented")
}

// Put is the wrapper function for an HTTP "PUT" request. This will create a
// new PUT request with a body composed of the contents of the io.Reader
// passed in, and the specified headers. The header map can be set to nil if no
// additional headers are required. If a nil ReadCloser is passed in, this will
// create an empty Put body which is allowed. This will return the full HTTP
// Response from the server, unaltered. This function returns an error and nil
// response on an HTTP StatusCode which is outside the 200 block.
func (C *SimpleClient) Put(URL *url.URL, Contents io.Reader, Headers map[string][]string) (*http.Response, error) {

	// Set the scheme, allowing for this to be missing, and be able to be defined by the internal TLS state of the Client.
	C.setScheme(URL)

	// Create the request
	req, err := NewRequest(http.MethodPut, URL, Headers, Contents)
	if err != nil {
		return nil, err
	}

	// Perform the request
	return C.Do(req)
}

// Delete is the wrapper function for an HTTP "DELETE" request. This will
// create a new DELETE request with an empty body, and the specified headers.
// The header map can be set to nil if no additional headers are required.
// This will ONLY any errors encountered, and no HTTP Response.
// The internal HTTP Response from the server will be safely closed by this function.
func (C *SimpleClient) Delete(URL *url.URL, Headers map[string][]string) (*http.Response, error) {

	// Set the scheme, allowing for this to be missing, and be able to be defined by the internal TLS state of the Client.
	C.setScheme(URL)

	// Create the request
	req, err := NewRequest(http.MethodDelete, URL, Headers, nil)
	if err != nil {
		return nil, err
	}

	// Perform the request
	resp, err := C.Do(req)
	switch err {
	case nil:
		defer resp.Body.Close()
		return resp, nil
	case ErrInvalidStatusCode:
		defer resp.Body.Close()
		return resp, err
	default:
		return nil, err
	}
}

// Patch is the wrapper function for an HTTP "PATCH" request. This will create
// a new PATCH request with a body composed of the contents of the io.Reader
// passed in, and the specified headers. The header map can be set to nil if no
// additional headers are required. If a nil ReadCloser is passed in, this will
// create an empty Patch body which is allowed. This will return the full HTTP
// Response from the server, unaltered. This function returns an error and nil
// response on an HTTP StatusCode which is outside the 200 block.
func (C *SimpleClient) Patch(URL *url.URL, Contents io.Reader, Headers map[string][]string) (*http.Response, error) {

	// Set the scheme, allowing for this to be missing, and be able to be defined by the internal TLS state of the Client.
	C.setScheme(URL)

	// Create the request
	req, err := NewRequest(http.MethodPatch, URL, Headers, Contents)
	if err != nil {
		return nil, err
	}

	// Perform the request
	return C.Do(req)
}

// Connect is the wrapper function for an HTTP "CONNECT" request.
//
// NOTE: This is not yet implemented, and does not appear to be supported by
// the standard library http package.
func (C *SimpleClient) Connect(URL *url.URL, Headers map[string][]string) error {
	return errors.New("Method CONNECT not yet implemented")
}

// Options is the wrapper function for an HTTP "OPTIONS" request. This will
// create a new OPTIONS request with an empty body, and the specified headers.
// The header map can be set to nil if no additional headers are required.
// This will return the full HTTP Response from the server, unaltered.
// This function returns an error and nil response on an HTTP StatusCode
//  which is outside the 200 block.
func (C *SimpleClient) Options(URL *url.URL, Headers map[string][]string) (*http.Response, error) {

	// Set the scheme, allowing for this to be missing, and be able to be defined by the internal TLS state of the Client.
	C.setScheme(URL)

	// Create the request
	req, err := NewRequest(http.MethodOptions, URL, Headers, nil)
	if err != nil {
		return nil, err
	}

	// Perform the request
	return C.Do(req)
}

// Trace is the wrapper function for an HTTP "TRACE" request. This will create
// a new TRACE request with an empty body, and the specified headers.
// The header map can be set to nil if no additional headers are required.
// This will return the full HTTP Response from the server, unaltered.
// This function returns an error and nil response on an HTTP StatusCode
// which is outside the 200 block.
func (C *SimpleClient) Trace(URL *url.URL, Headers map[string][]string) (*http.Response, error) {

	// Set the scheme, allowing for this to be missing, and be able to be defined by the internal TLS state of the Client.
	C.setScheme(URL)

	// Create the request
	req, err := NewRequest(http.MethodTrace, URL, Headers, nil)
	if err != nil {
		return nil, err
	}

	// Perform the request
	return C.Do(req)
}

// Do is the wrapper function for a generic pre-generated HTTP request.
//
// This is the generic underlying call used by the rest of this library.
//
// This will perform no alterations to the provided request, and no alterations
// to the returned Response. This function returns ErrInvalidStatusCode and the
// full response on an HTTP StatusCode which is outside the 200 block.
// This is still a meaningful response, but a helpful error to quickly
// disambiguate errors of transport versus errors of action.
//
// This function is extended by the use of a TLSRetryPolicy within the
// SimpleClient. This allows a client to attempt to handle HTTP/HTTPS mismatch
// errors automatically by upgrading/downgrading as necessary.
func (C *SimpleClient) Do(req *http.Request) (*http.Response, error) {

	// Set the scheme, allowing for this to be missing, and be able to be defined by the internal TLS state of the Client.
	C.setScheme(req.URL)

	resp, err := C.client.Do(req)
	if err != nil {
		return C.shouldDowngrade(req, resp, err)
	}

	if 200 <= resp.StatusCode && resp.StatusCode < 300 {
		return resp, nil
	}

	return C.shouldUpgrade(req, resp, ErrInvalidStatusCode)
}

// Check if the client should downgrade to HTTP and attempt to perform the request again.
func (C *SimpleClient) shouldDowngrade(req *http.Request, resp *http.Response, respErr error) (*http.Response, error) {

	// If the Client Retry policy indicates either no retries, or only upgrades, don't attempt anything
	if C.policy == NoRetry ||
		C.policy == UpgradeNoReset ||
		C.policy == UpgradeWithReset {
		return resp, respErr
	}

	// Check if we have a standard golang http package error indicating an HTTP response for HTTPS client.
	if strings.Contains(respErr.Error(), "http: server gave HTTP response to HTTPS client") {
		return C.retryWithDowngrade(req, resp)
	}

	// Perform more, or other checks...

	return resp, respErr
}

// Check if the client should try to upgrade to HTTPS and attempt to perform the request again.
func (C *SimpleClient) shouldUpgrade(req *http.Request, resp *http.Response, respErr error) (*http.Response, error) {

	if C.policy == NoRetry ||
		C.policy == DowngradeNoReset ||
		C.policy == DowngradeWithReset {
		return resp, respErr
	}

	// if resp.StatusCode == http.StatusBadRequest

	return resp, respErr
}

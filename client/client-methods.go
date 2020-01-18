package client

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
)

// ErrInvalidStatusCode references the Error returned by any request which succeeds in communicating with the server, but does not return a 2xx level response code.
var ErrInvalidStatusCode = errors.New("Invalid status code - Expected 2xx")

// Get is the wrapper function for an HTTP "GET" request. This will create a GET request with an empty body, and the specified headers. The header map can be set to nil if no additional headers are required. This function returns an error and nil response on an HTTP StatusCode which is outside the 200 block.
func (C *SimpleClient) Get(URL *url.URL, Headers map[string][]string) (*http.Response, error) {

	// Create the request
	req, err := NewRequest(http.MethodGet, URL, Headers, nil)
	if err != nil {
		return nil, err
	}

	// Perform the request
	return C.Do(req)
}

// Head is the wrapper function for an HTTP "HEAD" request. This will create a new HEAD request with an empty body and the specified headers. The header map can be set to nil if no additional headers are required. This will ONLY return the HTTP Response Header map from the server. The overall Response Body (if it exists) will be closed by this function. This function returns an error and nil Header on an HTTP StatusCode which is outside the 200 block.
func (C *SimpleClient) Head(URL *url.URL, Headers map[string][]string) (*http.Response, error) {

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

// Post is the wrapper function for an HTTP "POST" request. This will create a new POST request with a body composed of the contents of the io.Reader passed in, and the specified headers. The header map can be set to nil if no additional headers are required. If a nil ReadCloser is passed in, this will create an empty Post body which is allowed. This will return the full HTTP Response from the server, unaltered. This function returns an error and nil response on an HTTP StatusCode which is outside the 200 block.
//
// This function "may" support MultiPart POST requests, by way of io.Pipes and multipart.Writers, but this has not been tested, and multipart Post is planned to be an explicit function in the future.
func (C *SimpleClient) Post(URL *url.URL, Contents io.Reader, Headers map[string][]string) (*http.Response, error) {

	// Create the request
	req, err := NewRequest(http.MethodPost, URL, Headers, Contents)
	if err != nil {
		return nil, err
	}

	// Perform the request
	return C.Do(req)
}

// PostMultipart is the wrapper function for an HTTP "POST" request with a MultiPart Body. This will create a new POST request with a body composed of the contents of the multipart.Reader passed in, and the specified headers. The header map can be set to nil if no additional headers are required. If a nil multipart.Reader is passed in, this will create an empty Post body which is allowed. This will return the full HTTP Response from the server, unaltered. This function returns an error and nil response on an HTTP StatusCode which is outside the 200 block.
//
// (Not Yet Implemented)
func (C *SimpleClient) PostMultipart(URL *url.URL, Contents multipart.Reader, Headers map[string][]string) (*http.Response, error) {
	return nil, errors.New("Method POST-MULTIPART not yet implemented")
}

// Put is the wrapper function for an HTTP "PUT" request. This will create a new PUT request with a body composed of the contents of the io.Reader passed in, and the specified headers. The header map can be set to nil if no additional headers are required. If a nil ReadCloser is passed in, this will create an empty Put body which is allowed. This will return the full HTTP Response from the server, unaltered. This function returns an error and nil response on an HTTP StatusCode which is outside the 200 block.
func (C *SimpleClient) Put(URL *url.URL, Contents io.Reader, Headers map[string][]string) (*http.Response, error) {

	// Create the request
	req, err := NewRequest(http.MethodPut, URL, Headers, Contents)
	if err != nil {
		return nil, err
	}

	// Perform the request
	return C.Do(req)
}

// Delete is the wrapper function for an HTTP "DELETE" request. This will create a new DELETE request with an empty body, and the specified headers. The header map can be set to nil if no additional headers are required.
// This will return ONLY an error, and no HTTP Response components. The internal HTTP Response from the server will be safely closed by this function.
func (C *SimpleClient) Delete(URL *url.URL, Headers map[string][]string) (*http.Response, error) {

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

// Patch is the wrapper function for an HTTP "PATCH" request. This will create a new PATCH request with a body composed of the contents of the io.Reader passed in, and the specified headers. The header map can be set to nil if no additional headers are required. If a nil ReadCloser is passed in, this will create an empty Patch body which is allowed. This will return the full HTTP Response from the server, unaltered. This function returns an error and nil response on an HTTP StatusCode which is outside the 200 block.
func (C *SimpleClient) Patch(URL *url.URL, Contents io.Reader, Headers map[string][]string) (*http.Response, error) {

	// Create the request
	req, err := NewRequest(http.MethodPatch, URL, Headers, Contents)
	if err != nil {
		return nil, err
	}

	// Perform the request
	return C.Do(req)
}

// Connect is the wrapper function for an HTTP "CONNECT" request. (Not yet Implemented)
func (C *SimpleClient) Connect(URL *url.URL, Headers map[string][]string) error {
	return errors.New("Method CONNECT not yet implemented")
}

// Options is the wrapper function for an HTTP "OPTIONS" request. This will create a new OPTIONS request with an empty body, and the specified headers. The header map can be set to nil if no additional headers are required. This will return the full HTTP Response from the server, unaltered. This function returns an error and nil response on an HTTP StatusCode which is outside the 200 block.
func (C *SimpleClient) Options(URL *url.URL, Headers map[string][]string) (*http.Response, error) {

	// Create the request
	req, err := NewRequest(http.MethodOptions, URL, Headers, nil)
	if err != nil {
		return nil, err
	}

	// Perform the request
	return C.Do(req)
}

// Trace is the wrapper function for an HTTP "TRACE" request. This will create a new TRACE request with an empty body, and the specified headers. The header map can be set to nil if no additional headers are required. This will return the full HTTP Response from the server, unaltered. This function returns an error and nil response on an HTTP StatusCode which is outside the 200 block.
func (C *SimpleClient) Trace(URL *url.URL, Headers map[string][]string) (*http.Response, error) {

	// Create the request
	req, err := NewRequest(http.MethodTrace, URL, Headers, nil)
	if err != nil {
		return nil, err
	}

	// Perform the request
	return C.Do(req)
}

// Do is the wrapper function for a generic pre-generated HTTP request.
// This is the generic underlying call used by the rest of this library
//  (and reflects similarly to how HTTP Requests are handled in the standard library).
// This will perform no alterations to the provided request, and no alterations to the returned Response.
// This function ErrInvalidStatusCode and the full response on an HTTP StatusCode which is outside the 200 block.
// This is still a meaningful response, but a helpful error to quickly disambiguate errors of transport versus errors of action.
func (C *SimpleClient) Do(req *http.Request) (*http.Response, error) {
	resp, err := C.client.Do(req)
	if err != nil {
		return nil, err
	}

	if 200 <= resp.StatusCode && resp.StatusCode < 300 {
		return resp, nil
	}

	return resp, ErrInvalidStatusCode
}

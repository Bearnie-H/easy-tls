package client

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
)

// GetContext is the wrapper function for an HTTP "GET" request. This will create a
// GET request with an empty body and the specified headers. The header map can
// be set to nil if no additional headers are required. This function returns
// an error and nil response on an HTTP StatusCode which is outside the 200
// block.
func (C *SimpleClient) GetContext(ctx context.Context, URL string, Headers map[string][]string) (*http.Response, error) {

	// Create the request
	req, err := NewRequestWithContext(ctx, http.MethodGet, URL, Headers, nil)
	if err != nil {
		return nil, err
	}

	// Perform the request
	return C.Do(req)
}

// HeadContext is the wrapper function for an HTTP "HEAD" request. This will create a
// new HEAD request with an empty body and the specified headers. The header
// map can be set to nil if no additional headers are required. This will ONLY
// return the HTTP Response Header map from the server. The overall Response
// Body (if it exists) will be closed by this function. This function returns
// an error and nil Header on an HTTP StatusCode which is outside the 200
// block.
func (C *SimpleClient) HeadContext(ctx context.Context, URL string, Headers map[string][]string) (*http.Response, error) {

	// Create the request
	req, err := NewRequestWithContext(ctx, http.MethodHead, URL, Headers, nil)
	if err != nil {
		return nil, err
	}

	// Perform the request
	return C.Do(req)
}

// PostContext is the wrapper function for an HTTP "POST" request. This will create a
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
func (C *SimpleClient) PostContext(ctx context.Context, URL string, Contents io.Reader, Headers map[string][]string) (*http.Response, error) {

	// Create the request
	req, err := NewRequestWithContext(ctx, http.MethodPost, URL, Headers, Contents)
	if err != nil {
		return nil, err
	}

	// Perform the request
	return C.Do(req)
}

// PostMultipartContext is the wrapper function for an HTTP "POST" request with a
// MultiPart Body. This will create a new POST request with a body composed of
// the contents of the multipart.Reader passed in, and the specified headers.
// The header map can be set to nil if no additional headers are required. If
// a nil multipart.Reader is passed in, this will create an empty Post body
// which is allowed. This will return the full HTTP Response from the server
// unaltered. This function returns an error and nil response on an HTTP
// StatusCode which is outside the 200 block.
//
// NOTE: This has not yet been implemented.
func (C *SimpleClient) PostMultipartContext(ctx context.Context, URL string, Contents multipart.Reader, Headers map[string][]string) (*http.Response, error) {
	return nil, errors.New("Method POST-MULTIPART not yet implemented")
}

// PutContext is the wrapper function for an HTTP "PUT" request. This will create a
// new PUT request with a body composed of the contents of the io.Reader
// passed in, and the specified headers. The header map can be set to nil if no
// additional headers are required. If a nil ReadCloser is passed in, this will
// create an empty Put body which is allowed. This will return the full HTTP
// Response from the server, unaltered. This function returns an error and nil
// response on an HTTP StatusCode which is outside the 200 block.
func (C *SimpleClient) PutContext(ctx context.Context, URL string, Contents io.Reader, Headers map[string][]string) (*http.Response, error) {

	// Create the request
	req, err := NewRequestWithContext(ctx, http.MethodPut, URL, Headers, Contents)
	if err != nil {
		return nil, err
	}

	// Perform the request
	return C.Do(req)
}

// DeleteContext is the wrapper function for an HTTP "DELETE" request. This will
// create a new DELETE request with an empty body, and the specified headers.
// The header map can be set to nil if no additional headers are required.
// This will ONLY any errors encountered, and no HTTP Response.
// The internal HTTP Response from the server will be safely closed by this function.
func (C *SimpleClient) DeleteContext(ctx context.Context, URL string, Headers map[string][]string) (*http.Response, error) {

	// Create the request
	req, err := NewRequestWithContext(ctx, http.MethodDelete, URL, Headers, nil)
	if err != nil {
		return nil, err
	}

	// Perform the request
	return C.Do(req)
}

// PatchContext is the wrapper function for an HTTP "PATCH" request. This will create
// a new PATCH request with a body composed of the contents of the io.Reader
// passed in, and the specified headers. The header map can be set to nil if no
// additional headers are required. If a nil ReadCloser is passed in, this will
// create an empty Patch body which is allowed. This will return the full HTTP
// Response from the server, unaltered. This function returns an error and nil
// response on an HTTP StatusCode which is outside the 200 block.
func (C *SimpleClient) PatchContext(ctx context.Context, URL string, Contents io.Reader, Headers map[string][]string) (*http.Response, error) {

	// Create the request
	req, err := NewRequestWithContext(ctx, http.MethodPatch, URL, Headers, Contents)
	if err != nil {
		return nil, err
	}

	// Perform the request
	return C.Do(req)
}

// OptionsContext is the wrapper function for an HTTP "OPTIONS" request. This will
// create a new OPTIONS request with an empty body, and the specified headers.
// The header map can be set to nil if no additional headers are required.
// This will return the full HTTP Response from the server, unaltered.
// This function returns an error and nil response on an HTTP StatusCode
//  which is outside the 200 block.
func (C *SimpleClient) OptionsContext(ctx context.Context, URL string, Headers map[string][]string) (*http.Response, error) {

	// Create the request
	req, err := NewRequestWithContext(ctx, http.MethodOptions, URL, Headers, nil)
	if err != nil {
		return nil, err
	}

	// Perform the request
	return C.Do(req)
}

// TraceContext is the wrapper function for an HTTP "TRACE" request. This will create
// a new TRACE request with an empty body, and the specified headers.
// The header map can be set to nil if no additional headers are required.
// This will return the full HTTP Response from the server, unaltered.
// This function returns an error and nil response on an HTTP StatusCode
// which is outside the 200 block.
func (C *SimpleClient) TraceContext(ctx context.Context, URL string, Headers map[string][]string) (*http.Response, error) {

	// Create the request
	req, err := NewRequestWithContext(ctx, http.MethodTrace, URL, Headers, nil)
	if err != nil {
		return nil, err
	}

	// Perform the request
	return C.Do(req)
}

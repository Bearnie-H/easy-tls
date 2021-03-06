package client

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
)

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
func (C *SimpleClient) Get(URL string, Headers map[string][]string) (*http.Response, error) {
	return C.GetContext(context.Background(), URL, Headers)
}

// Head is the wrapper function for an HTTP "HEAD" request. This will create a
// new HEAD request with an empty body and the specified headers. The header
// map can be set to nil if no additional headers are required. This will ONLY
// return the HTTP Response Header map from the server. The overall Response
// Body (if it exists) will be closed by this function. This function returns
// an error and nil Header on an HTTP StatusCode which is outside the 200
// block.
func (C *SimpleClient) Head(URL string, Headers map[string][]string) (*http.Response, error) {
	return C.HeadContext(context.Background(), URL, Headers)
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
func (C *SimpleClient) Post(URL string, Contents io.Reader, Headers map[string][]string) (*http.Response, error) {
	return C.PostContext(context.Background(), URL, Contents, Headers)
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
func (C *SimpleClient) PostMultipart(URL string, Contents multipart.Reader, Headers map[string][]string) (*http.Response, error) {
	return C.PostMultipartContext(context.Background(), URL, Contents, Headers)
}

// Put is the wrapper function for an HTTP "PUT" request. This will create a
// new PUT request with a body composed of the contents of the io.Reader
// passed in, and the specified headers. The header map can be set to nil if no
// additional headers are required. If a nil ReadCloser is passed in, this will
// create an empty Put body which is allowed. This will return the full HTTP
// Response from the server, unaltered. This function returns an error and nil
// response on an HTTP StatusCode which is outside the 200 block.
func (C *SimpleClient) Put(URL string, Contents io.Reader, Headers map[string][]string) (*http.Response, error) {
	return C.PutContext(context.Background(), URL, Contents, Headers)
}

// Delete is the wrapper function for an HTTP "DELETE" request. This will
// create a new DELETE request with an empty body, and the specified headers.
// The header map can be set to nil if no additional headers are required.
// This will ONLY any errors encountered, and no HTTP Response.
// The internal HTTP Response from the server will be safely closed by this function.
func (C *SimpleClient) Delete(URL string, Headers map[string][]string) (*http.Response, error) {
	return C.DeleteContext(context.Background(), URL, Headers)
}

// Patch is the wrapper function for an HTTP "PATCH" request. This will create
// a new PATCH request with a body composed of the contents of the io.Reader
// passed in, and the specified headers. The header map can be set to nil if no
// additional headers are required. If a nil ReadCloser is passed in, this will
// create an empty Patch body which is allowed. This will return the full HTTP
// Response from the server, unaltered. This function returns an error and nil
// response on an HTTP StatusCode which is outside the 200 block.
func (C *SimpleClient) Patch(URL string, Contents io.Reader, Headers map[string][]string) (*http.Response, error) {
	return C.PatchContext(context.Background(), URL, Contents, Headers)
}

// Options is the wrapper function for an HTTP "OPTIONS" request. This will
// create a new OPTIONS request with an empty body, and the specified headers.
// The header map can be set to nil if no additional headers are required.
// This will return the full HTTP Response from the server, unaltered.
// This function returns an error and nil response on an HTTP StatusCode
//  which is outside the 200 block.
func (C *SimpleClient) Options(URL string, Headers map[string][]string) (*http.Response, error) {
	return C.OptionsContext(context.Background(), URL, Headers)
}

// Trace is the wrapper function for an HTTP "TRACE" request. This will create
// a new TRACE request with an empty body, and the specified headers.
// The header map can be set to nil if no additional headers are required.
// This will return the full HTTP Response from the server, unaltered.
// This function returns an error and nil response on an HTTP StatusCode
// which is outside the 200 block.
func (C *SimpleClient) Trace(URL string, Headers map[string][]string) (*http.Response, error) {
	return C.TraceContext(context.Background(), URL, Headers)
}

// Do is the wrapper function for a generic pre-generated HTTP request.
//
// This is the generic underlying call used by the rest of this library.
func (C *SimpleClient) Do(req *http.Request) (*http.Response, error) {
	C.setScheme(req.URL)
	return C.Client.Do(req)
}

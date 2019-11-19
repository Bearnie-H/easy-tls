package easytls

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Get represents the abstraction of the HTTP Get request, accounting for creating the request, setting headers, and asserting a valid status code.  Closing the response body is the responsibility of the caller.
func (C *SimpleClient) Get(URL *url.URL, Headers map[string][]string) (*http.Response, error) {

	// Create the request
	req, err := newRequest(http.MethodGet, URL, Headers, nil)
	if err != nil {
		return nil, err
	}

	// Perform the request
	return C.Do(req)
}

// Head represents the abstraction of the HTTP Head request, accounting for creating the request, setting headers, and asserting a valid status code.  Closing the response body is the responsibility of this function, as Head only returns the headers.
func (C *SimpleClient) Head(URL *url.URL, Headers map[string][]string) (http.Header, error) {

	// Create the request
	req, err := newRequest(http.MethodHead, URL, Headers, nil)
	if err != nil {
		return nil, err
	}

	// Perform the request
	resp, err := C.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return resp.Header, nil
}

// Post represents the abstraction of the HTTP Post request, accounting for creating the request, setting headers, and asserting a valid status code.  Closing the response body is the responsibility of the caller.
func (C *SimpleClient) Post(URL *url.URL, Contents io.ReadCloser, Headers map[string][]string) (*http.Response, error) {

	// Create the request
	req, err := newRequest(http.MethodPost, URL, Headers, Contents)
	if err != nil {
		return nil, err
	}

	// Perform the request
	return C.Do(req)
}

// Put represents the abstraction of the HTTP Put request, accounting for creating the request, setting headers, and asserting a valid status code.  Closing the response body is the responsibility of the caller.
func (C *SimpleClient) Put(URL *url.URL, Contents io.ReadCloser, Headers map[string][]string) (*http.Response, error) {

	// Create the request
	req, err := newRequest(http.MethodPut, URL, Headers, Contents)
	if err != nil {
		return nil, err
	}

	// Perform the request
	return C.Do(req)
}

// Delete represents the abstraction of the HTTP Delete request, accounting for creating the request, setting headers, and asserting a valid status code.  Closing the response body is the responsibility of this function.
func (C *SimpleClient) Delete(URL *url.URL, Headers map[string][]string) error {

	// Create the request
	req, err := newRequest(http.MethodDelete, URL, Headers, nil)
	if err != nil {
		return err
	}

	// Perform the request
	resp, err := C.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

// Patch represents the abstraction of the HTTP Patch request, accounting for creating the request, setting headers, and asserting a valid status code.  Closing the response body is the responsibility of this function.
func (C *SimpleClient) Patch(URL *url.URL, Contents io.ReadCloser, Headers map[string][]string) (*http.Response, error) {

	// Create the request
	req, err := newRequest(http.MethodPatch, URL, Headers, Contents)
	if err != nil {
		return nil, err
	}

	// Perform the request
	return C.Do(req)
}

// Connect will (Not yet implemented)
func (C *SimpleClient) Connect(URL *url.URL, Headers map[string][]string) error {
	return errors.New("Method CONNECT not yet implemented")
}

// Options represents the abstraction of the HTTP Options request, accounting for creating the request, setting headers, and asserting a valid status code.  Closing the response body is the responsibility of the caller.
func (C *SimpleClient) Options(URL *url.URL, Headers map[string][]string) (*http.Response, error) {

	// Create the request
	req, err := newRequest(http.MethodOptions, URL, Headers, nil)
	if err != nil {
		return nil, err
	}

	// Perform the request
	return C.Do(req)
}

// Trace represents the abstraction of the HTTP Trace request, accounting for creating the request, setting headers, and asserting a valid status code.  Closing the response body is the responsibility of the caller.
func (C *SimpleClient) Trace(URL *url.URL, Headers map[string][]string) (*http.Response, error) {

	// Create the request
	req, err := newRequest(http.MethodTrace, URL, Headers, nil)
	if err != nil {
		return nil, err
	}

	// Perform the request
	return C.Do(req)
}

// Do will perform a single pre-formatted request.
func (C *SimpleClient) Do(req *http.Request) (*http.Response, error) {
	resp, err := C.client.Do(req)
	if err != nil {
		return nil, err
	}

	if 200 <= resp.StatusCode && resp.StatusCode < 300 {
		return resp, nil
	}

	defer resp.Body.Close()
	return nil, fmt.Errorf("Invalid status code - expected 2xx, got %d (%s)", resp.StatusCode, resp.Status)
}

func newRequest(Method string, URL *url.URL, Headers map[string][]string, Contents io.ReadCloser) (*http.Request, error) {
	req, err := http.NewRequest(Method, URL.String(), Contents)
	if err != nil {
		return nil, err
	}

	for k, vs := range Headers {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}

	return req, nil
}

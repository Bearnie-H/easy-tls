package easytls

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Get represents the abstraction of the HTTP Get request, accounting for creating the request, setting headers, and asserting a valid status code.  Closing the response body is the responsibility of the caller.
func (C *SimpleClient) Get(URL *url.URL, Headers map[string]string) (*http.Response, error) {

	req, err := http.NewRequest(http.MethodGet, URL.String(), nil)
	if err != nil {
		return nil, err
	}

	// Set the given headers
	for k, v := range Headers {
		req.Header.Set(k, v)
	}

	// Perform the request
	resp, err := C.client.Do(req)
	if err != nil {
		return nil, err
	}

	/// If the status code is OK, return
	if 200 <= resp.StatusCode && resp.StatusCode < 300 {
		return resp, nil
	}

	// Otherwise, attempt to close whatever body we got, and return an error.
	resp.Body.Close()
	return nil, fmt.Errorf("Invalid status code - expected 2xx, got %d (%s)", resp.StatusCode, resp.Status)
}

// Head represents the abstraction of the HTTP Head request, accounting for creating the request, setting headers, and asserting a valid status code.  Closing the response body is the responsibility of this function, as Head only returns the headers.
func (C *SimpleClient) Head(URL *url.URL, Headers map[string]string) (http.Header, error) {

	req, err := http.NewRequest(http.MethodHead, URL.String(), nil)
	if err != nil {
		return nil, err
	}

	// Set the given headers
	for k, v := range Headers {
		req.Header.Set(k, v)
	}

	// Perform the request
	resp, err := C.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	/// If the status code is OK, return
	if 200 <= resp.StatusCode && resp.StatusCode < 300 {
		return resp.Header, nil
	}

	// Otherwise, attempt to close whatever body we got, and return an error.
	return nil, fmt.Errorf("Invalid status code - expected 2xx, got %d (%s)", resp.StatusCode, resp.Status)
}

// Post represents the abstraction of the HTTP Post request, accounting for creating the request, setting headers, and asserting a valid status code.  Closing the response body is the responsibility of the caller.
func (C *SimpleClient) Post(URL *url.URL, Contents io.ReadCloser, Headers map[string]string) (*http.Response, error) {

	req, err := http.NewRequest(http.MethodPost, URL.String(), Contents)
	if err != nil {
		return nil, err
	}

	// Set the given headers
	for k, v := range Headers {
		req.Header.Set(k, v)
	}

	// Perform the request
	resp, err := C.client.Do(req)
	if err != nil {
		return nil, err
	}

	/// If the status code is OK, return
	if 200 <= resp.StatusCode && resp.StatusCode < 300 {
		return resp, nil
	}

	// Otherwise, attempt to close whatever body we got, and return an error.
	resp.Body.Close()
	return nil, fmt.Errorf("Invalid status code - expected 2xx, got %d (%s)", resp.StatusCode, resp.Status)
}

// Put represents the abstraction of the HTTP Put request, accounting for creating the request, setting headers, and asserting a valid status code.  Closing the response body is the responsibility of the caller.
func (C *SimpleClient) Put(URL *url.URL, Contents io.ReadCloser, Headers map[string]string) (*http.Response, error) {

	req, err := http.NewRequest(http.MethodPut, URL.String(), Contents)
	if err != nil {
		return nil, err
	}

	// Set the given headers
	for k, v := range Headers {
		req.Header.Set(k, v)
	}

	// Perform the request
	resp, err := C.client.Do(req)
	if err != nil {
		return nil, err
	}

	/// If the status code is OK, return
	if 200 <= resp.StatusCode && resp.StatusCode < 300 {
		return resp, nil
	}

	defer resp.Body.Close()
	// Otherwise, attempt to close whatever body we got, and return an error.
	return nil, fmt.Errorf("Invalid status code - expected 2xx, got %d (%s)", resp.StatusCode, resp.Status)
}

// Delete represents the abstraction of the HTTP Delete request, accounting for creating the request, setting headers, and asserting a valid status code.  Closing the response body is the responsibility of this function.
func (C *SimpleClient) Delete(URL *url.URL, Headers map[string]string) error {

	req, err := http.NewRequest(http.MethodDelete, URL.String(), nil)
	if err != nil {
		return err
	}

	// Set the given headers
	for k, v := range Headers {
		req.Header.Set(k, v)
	}

	// Perform the request
	resp, err := C.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	/// If the status code is OK, return
	if 200 <= resp.StatusCode && resp.StatusCode < 300 {
		return nil
	}

	// Otherwise, attempt to close whatever body we got, and return an error.
	return fmt.Errorf("Invalid status code - expected 2xx, got %d (%s)", resp.StatusCode, resp.Status)
}

// Patch represents the abstraction of the HTTP Patch request, accounting for creating the request, setting headers, and asserting a valid status code.  Closing the response body is the responsibility of this function.
func (C *SimpleClient) Patch(URL *url.URL, Contents io.ReadCloser, Headers map[string]string) (*http.Response, error) {

	req, err := http.NewRequest(http.MethodPatch, URL.String(), Contents)
	if err != nil {
		return nil, err
	}

	// Set the given headers
	for k, v := range Headers {
		req.Header.Set(k, v)
	}

	// Perform the request
	resp, err := C.client.Do(req)
	if err != nil {
		return nil, err
	}

	/// If the status code is OK, return
	if 200 <= resp.StatusCode && resp.StatusCode < 300 {
		return resp, nil
	}

	defer resp.Body.Close()
	// Otherwise, attempt to close whatever body we got, and return an error.
	return nil, fmt.Errorf("Invalid status code - expected 2xx, got %d (%s)", resp.StatusCode, resp.Status)
}

// Connect will (Not yet implemented)
func (C *SimpleClient) Connect(URL *url.URL, Headers map[string]string) error {
	return errors.New("Method CONNECT not yet implemented")
}

// Options represents the abstraction of the HTTP Options request, accounting for creating the request, setting headers, and asserting a valid status code.  Closing the response body is the responsibility of the caller.
func (C *SimpleClient) Options(URL *url.URL, Headers map[string]string) (*http.Response, error) {

	req, err := http.NewRequest(http.MethodOptions, URL.String(), nil)
	if err != nil {
		return nil, err
	}

	// Set the given headers
	for k, v := range Headers {
		req.Header.Set(k, v)
	}

	// Perform the request
	resp, err := C.client.Do(req)
	if err != nil {
		return nil, err
	}

	/// If the status code is OK, return
	if 200 <= resp.StatusCode && resp.StatusCode < 300 {
		return resp, nil
	}

	// Otherwise, attempt to close whatever body we got, and return an error.
	defer resp.Body.Close()
	return nil, fmt.Errorf("Invalid status code - expected 2xx, got %d (%s)", resp.StatusCode, resp.Status)
}

// Trace represents the abstraction of the HTTP Trace request, accounting for creating the request, setting headers, and asserting a valid status code.  Closing the response body is the responsibility of the caller.
func (C *SimpleClient) Trace(URL *url.URL, Headers map[string]string) (*http.Response, error) {

	req, err := http.NewRequest(http.MethodTrace, URL.String(), nil)
	if err != nil {
		return nil, err
	}

	// Set the given headers
	for k, v := range Headers {
		req.Header.Set(k, v)
	}

	// Perform the request
	resp, err := C.client.Do(req)
	if err != nil {
		return nil, err
	}

	/// If the status code is OK, return
	if 200 <= resp.StatusCode && resp.StatusCode < 300 {
		return resp, nil
	}

	// Otherwise, attempt to close whatever body we got, and return an error.
	defer resp.Body.Close()
	return nil, fmt.Errorf("Invalid status code - expected 2xx, got %d (%s)", resp.StatusCode, resp.Status)
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

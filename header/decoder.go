package header

import "net/http"

// Decoder will implement the necessary functionality for parsing an http.Header into a Go struct.
type Decoder struct {
	v interface{}
	h http.Header
}

// NewDecoder will create and initialize a decoder, ready to write an http.Header into a Go struct.
// This will decode into the value pointed to by v.
func NewDecoder(v interface{}) *Decoder {
	return &Decoder{v: v}
}

// Decode will actually decode the Header H.
func (D *Decoder) Decode(H http.Header) error {
	D.h = H
	return D.decode()
}

func (D *Decoder) decode() error {

	// ...

	return nil
}

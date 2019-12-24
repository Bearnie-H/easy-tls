package header

import "net/http"

// Encoder will implement the necessary functionality for parsing a Go struct into a useable http.Header
type Encoder struct {
	h http.Header
	v interface{}
}

// NewEncoder will create and initialize an encoder, ready to write a Go struct into an HTTP Header
func NewEncoder(H *http.Header) *Encoder {
	return &Encoder{h: *H}
}

// Encode will actually encode the struct v into the http.Header.
func (E *Encoder) Encode(v interface{}) error {
	E.v = v
	return E.encode()
}

func (E *Encoder) encode() error {

	// ...

	return nil
}

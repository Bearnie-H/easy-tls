// Package header implements utility functions for encoding Go structs into
// standard HTTP Headers.
//
// The primary utility provided is an API for encoding a (nearly) arbitrary Go
// struct into an http.Header, as well as the converse of filling in the fields
// of a (nearly) arbitrary Go struct from an http.Header. This API also
// includes a struct tag, to allow for customized encoding of struct fields as
// keys when encoding into an http.Header.
//
// There are limitations on the exact nature of what can be encoded and decoded
// by this package, based primarily on the limitations of an http.Header just
// being a map[string][]string at its core.
//
// The following types are fully supported:
//
//	int, []int, and all bit-specified types
//	float, []float, and all bit-specified types
//	string, []string
//	bool, []bool
//	struct
//
// Arrays of structs are not supported, and issues may arise if multiple
// structs are encoded where the Field Names are duplicated. Nested structs
// are supported, but with the restriction on Field Name uniqueness.
//
// This package works by converting the Field Name (or struct tag, if present)
// into the map key, and "stringifying" the Field Value to append to the map
// value array. When decoding, the reverse process is used, where the fields of
// the struct to decode into are used to find corresponding values in the map,
// and the map values are parsed back into the corresponding types.
//
package header

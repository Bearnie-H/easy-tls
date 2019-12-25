// Package header implements utility functions for working with HTTP Headers, and the necessary encoding/decoding of them.
//
// The primary utility provided is an API for encoding a (nearly) arbitrary Go struct into an http.Header,
//  as well as the converse of filling in the fields of a (nearly) arbitrary Go struct from an http.Header.
//
// Currently, the only types capable of being encoded and decoded with this API are:
//
//		int (and all bit-specified derived types, resolves to int64 internally)
//		bool
//		float (and all bit-specified derived types, resolves to float64 internally)
//		string
//		As well as slices of these types.
//
// Since this does not include structs, there is no current mechanism for converting nested structs with this API.
// This is unlikely be added, as nested structs do not exactly play nicely with the map[string][]string underlying structure,
//  and an approach like a MIME Multipart message with a JSON preamble is a more standard solution.
// If unsupported types are found, they are simply ignored.
package header

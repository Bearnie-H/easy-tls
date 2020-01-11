package header

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
)

const (
	// EasyTLSStructTag represents the Struct Tag key used by this package
	EasyTLSStructTag = `easytls`
)

// Encoder will implement the necessary functionality for parsing a Go struct into a useable http.Header
type Encoder struct {
	h http.Header
	v interface{}
}

// NewEncoder will create and initialize an encoder, ready to write a Go struct into an HTTP Header
func NewEncoder(H *http.Header) *Encoder {
	if H == nil {
		H = &http.Header{}
	}
	return &Encoder{h: *H}
}

// DefaultEncode will allow using a default encoding, returning an HTTP Header to be merged with the one to be sent.
func DefaultEncode(v interface{}) (http.Header, error) {
	enc := NewEncoder(&http.Header{})
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return enc.Header(), nil
}

// Header returns a copy of the underlying HTTP Header, as the Encoder currenty sees it.
func (E *Encoder) Header() http.Header {
	return E.h
}

// Encode will actually encode the struct v into the http.Header.
// This process must flatten any nested structs into a single layer, to match the underlying structure of the HTTP Headers.
func (E *Encoder) Encode(v interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("header - encode error - %s", r)
			}
		}
	}()
	if !reflect.ValueOf(v).IsValid() || v == nil {
		err = errors.New("encoder error: Invalid or nil interface provided")
		return
	}
	err = E.encode(v)
	return
}

func (E *Encoder) encode(v interface{}) error {

	InVal := reflect.ValueOf(v)
	InType := reflect.TypeOf(v)

	// Iterate over the fields of the struct to encode
	for i := 0; i < InVal.NumField(); i++ {
		FieldName := InType.Field(i).Name
		FieldType := InType.Field(i).Type.Kind()
		FieldValue := InVal.Field(i)
		FieldTag := InType.Field(i).Tag.Get(EasyTLSStructTag)
		if FieldTag == "-" {
			continue
		}
		if FieldTag != "" {
			FieldName = FieldTag
		}
		switch FieldType {
		case reflect.Bool:
			E.encodeBool(FieldName, FieldValue)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			E.encodeInt(FieldName, FieldValue)
		case reflect.Float32, reflect.Float64:
			E.encodeFloat(FieldName, FieldValue)
		case reflect.Array, reflect.Slice:
			if err := E.encodeSlice(FieldName, FieldValue); err != nil {
				return err
			}
		case reflect.String:
			E.encodeString(FieldName, FieldValue)
		case reflect.Struct:
			sub := FieldValue.Interface()
			if err := E.encode(sub); err != nil {
				return err
			}
		default:
		}
	}
	return nil
}

func (E *Encoder) encodeBool(Name string, Val reflect.Value) {
	if !Val.IsValid() {
		return
	}
	val := Val.Bool()
	E.h.Add(Name, fmt.Sprintf("%t", val))
}

func (E *Encoder) encodeInt(Name string, Val reflect.Value) {
	if !Val.IsValid() {
		return
	}
	val := Val.Int()
	E.h.Add(Name, fmt.Sprintf("%d", val))
}

func (E *Encoder) encodeFloat(Name string, Val reflect.Value) {
	if !Val.IsValid() {
		return
	}
	val := Val.Float()
	E.h.Add(Name, fmt.Sprintf("%g", val))
}

func (E *Encoder) encodeString(Name string, Val reflect.Value) {
	if !Val.IsValid() {
		return
	}
	val := Val.String()
	E.h.Add(Name, val)
}

func (E *Encoder) encodeSlice(Name string, Val reflect.Value) error {
	if !Val.IsValid() || Val.IsNil() {
		return nil
	}
	SliceKind := Val.Type().Elem().Kind()
	switch SliceKind {
	case reflect.Bool:
		for i := 0; i < Val.Len(); i++ {
			E.Header().Add(Name, fmt.Sprintf("%t", Val.Index(i).Bool()))
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		for i := 0; i < Val.Len(); i++ {
			E.Header().Add(Name, fmt.Sprintf("%d", Val.Index(i).Int()))
		}
	case reflect.Float32, reflect.Float64:
		for i := 0; i < Val.Len(); i++ {
			E.Header().Add(Name, fmt.Sprintf("%f", Val.Index(i).Float()))
		}
	case reflect.String:
		for i := 0; i < Val.Len(); i++ {
			E.Header().Add(Name, Val.Index(i).String())
		}
	default:
		return fmt.Errorf("encoder error: Unsupported slice type - %s", SliceKind.String())
	}

	return nil
}

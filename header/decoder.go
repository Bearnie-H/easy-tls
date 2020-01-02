package header

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
)

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

// DefaultDecode will allow using the default decoder, decoding the header into the value pointed to by v.
func DefaultDecode(H http.Header, v interface{}) error {
	dec := NewDecoder(v)
	return dec.Decode(H)
}

// Out will return a copy of the struct being filled in by this Decoder, exactly as it sees it at the time of calling.
func (D *Decoder) Out() interface{} {
	return D.v
}

// Decode will actually decode the Header H.
func (D *Decoder) Decode(H http.Header) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("header - encode error - %s", r)
			}
		}
	}()
	D.h = H
	err = D.decode()
	return
}

func (D *Decoder) decode() error {

	OutVal := reflect.ValueOf(D.v).Elem()
	OutType := reflect.TypeOf(D.v).Elem()

	// Iterate over the struct fields...
	for i := 0; i < OutVal.NumField(); i++ {
		FieldName := OutType.Field(i).Name
		FieldType := OutType.Field(i).Type.Kind()
		FieldValue := OutVal.Field(i)
		HeaderValue := D.h[http.CanonicalHeaderKey(FieldName)]
		switch FieldType {
		case reflect.Bool:
			v, err := D.decodeBool(HeaderValue)
			if err != nil {
				return err
			}
			FieldValue.SetBool(v)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			v, err := D.decodeInt(HeaderValue)
			if err != nil {
				return err
			}
			FieldValue.SetInt(int64(v))
		case reflect.Float32, reflect.Float64:
			v, err := D.decodeFloat(HeaderValue)
			if err != nil {
				return err
			}
			FieldValue.SetFloat(v)
		case reflect.Array, reflect.Slice:
			v, err := D.decodeSlice(HeaderValue, FieldValue.Type().Elem().Kind())
			if err != nil {
				return err
			}
			var V reflect.Value
			switch v.(type) {
			case []bool:
				V = reflect.MakeSlice(FieldValue.Type(), len(v.([]bool)), cap(v.([]bool)))
				for index, val := range v.([]bool) {
					V.Index(index).SetBool(val)
				}
			case []int:
				V = reflect.MakeSlice(FieldValue.Type(), len(v.([]int)), cap(v.([]int)))
				for index, val := range v.([]int) {
					V.Index(index).SetInt(int64(val))
				}
			case []float64:
				V = reflect.MakeSlice(FieldValue.Type(), len(v.([]float64)), cap(v.([]float64)))
				for index, val := range v.([]float64) {
					V.Index(index).SetFloat(val)
				}
			case []string:
				V = reflect.MakeSlice(FieldValue.Type(), len(v.([]string)), cap(v.([]string)))
				for index, val := range v.([]string) {
					V.Index(index).SetString(val)
				}
			default:
				V = reflect.MakeSlice(FieldValue.Type().Elem(), 0, 0)
			}
			FieldValue.Set(V)
		case reflect.String:
			v, err := D.decodeString(HeaderValue)
			if err != nil {
				return err
			}
			FieldValue.SetString(v)
		default:
		}
	}

	return nil
}

func (D *Decoder) decodeBool(HeaderValue []string) (bool, error) {
	if HeaderValue == nil || len(HeaderValue) == 0 {
		return false, nil
	}
	switch HeaderValue[0] {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("decoder error: Invalid boolean value - %s", HeaderValue[0])
	}
}

func (D *Decoder) decodeInt(HeaderValue []string) (int, error) {
	if HeaderValue == nil || len(HeaderValue) == 0 {
		return 0, nil
	}

	temp, err := strconv.ParseInt(HeaderValue[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("decoder error: Invalid integer value (%s) - %s", HeaderValue[0], err)
	}

	return int(temp), nil
}

func (D *Decoder) decodeFloat(HeaderValue []string) (float64, error) {
	if HeaderValue == nil || len(HeaderValue) == 0 {
		return 0, nil
	}

	temp, err := strconv.ParseFloat(HeaderValue[0], 64)
	if err != nil {
		return 0, fmt.Errorf("decoder error: Invalid float value (%s) - %s", HeaderValue[0], err)
	}

	return temp, nil
}

func (D *Decoder) decodeString(HeaderValue []string) (string, error) {
	if HeaderValue == nil || len(HeaderValue) == 0 {
		return "", nil
	}

	return HeaderValue[0], nil
}

func (D *Decoder) decodeSlice(HeaderValue []string, SliceKind reflect.Kind) (interface{}, error) {
	switch SliceKind {
	case reflect.Bool:
		s := []bool{}
		for _, x := range HeaderValue {
			v, err := D.decodeBool([]string{x})
			if err != nil {
				return nil, err
			}
			s = append(s, v)
		}
		return s, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		s := []int{}
		for _, x := range HeaderValue {
			v, err := D.decodeInt([]string{x})
			if err != nil {
				return nil, err
			}
			s = append(s, v)
		}
		return s, nil
	case reflect.Float32, reflect.Float64:
		s := []float64{}
		for _, x := range HeaderValue {
			v, err := D.decodeFloat([]string{x})
			if err != nil {
				return nil, err
			}
			s = append(s, v)
		}
		return s, nil
	case reflect.String:
		s := []string{}
		for _, x := range HeaderValue {
			v, err := D.decodeString([]string{x})
			if err != nil {
				return nil, err
			}
			s = append(s, v)
		}
		return s, nil
	default:
		return nil, fmt.Errorf("encoder error: Unsupported slice type - %s", SliceKind.String())
	}
}

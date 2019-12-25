package header

import (
	"log"
	"net/http"
	"testing"
)

// TestStruct represents a basic test struct containing all of the allowable Types within an HTTP Header
type TestStruct struct {
	IntTest         int
	BoolTest        bool
	StringTest      string
	FloatTest       float64
	BoolSliceTest   []bool
	IntSliceTest    []int
	StringSliceTest []string
	FloatSliceTest  []float64
}

func TestEncoder(t *testing.T) {

	H := &http.Header{}
	S := TestStruct{
		IntTest:         0,
		IntSliceTest:    []int{0, 1},
		BoolTest:        true,
		BoolSliceTest:   []bool{true, false},
		StringTest:      "Test",
		StringSliceTest: []string{"TEst", "TESSST"},
		FloatTest:       3.14,
		FloatSliceTest:  []float64{4.55, 420.2},
	}

	E := NewEncoder(H)

	if err := E.Encode(S); err != nil {
		log.Fatalln(err)
	}

	log.Printf("%+v", H)
}

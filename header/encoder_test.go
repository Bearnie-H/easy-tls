package header

import (
	"fmt"
	"log"
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

	S := TestStruct{
		IntTest:         238947234,
		IntSliceTest:    []int{0, 1},
		BoolTest:        true,
		BoolSliceTest:   []bool{true, false},
		StringTest:      "Test",
		StringSliceTest: []string{"TEst", "TESSST"},
		FloatTest:       3.14,
		FloatSliceTest:  []float64{4.55, 420.2},
	}

	fmt.Printf("Struct before any encoding:\n%+v\n", S)

	H, err := DefaultEncode(&S)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("Struct after encoding:\n%+v\nHeader constructed from Struct:\n%+v\n", S, H)
}

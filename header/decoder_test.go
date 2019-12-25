package header

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"testing"
)

func Test_Decoder(t *testing.T) {

	H := &http.Header{}
	S := TestStruct{
		IntTest:         420,
		IntSliceTest:    []int{69, 42069},
		BoolTest:        true,
		BoolSliceTest:   []bool{true, false, true, true, false},
		StringTest:      "Chungus - Big Chungus",
		StringSliceTest: []string{"This", "is", "a", "test", "of", "strings"},
		FloatTest:       math.Pi,
		FloatSliceTest:  []float64{math.SqrtPi, math.SqrtE},
	}

	E := NewEncoder(H)

	if err := E.Encode(S); err != nil {
		log.Fatalln(err)
	}

	S2 := TestStruct{}

	D := NewDecoder(&S2)

	if err := D.Decode(*H); err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("Pre-Encoding: %+v\n\nPost-Encoding: %+v\n", S, S2)
}

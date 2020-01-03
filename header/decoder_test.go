package header

import (
	"fmt"
	"log"
	"math"
	"testing"
)

type Foo struct {
	Bar string
	Baz int
}

type test2 struct {
	TestStruct
	Foo
}

func Test_Decoder(t *testing.T) {

	s := TestStruct{
		IntTest:         420,
		IntSliceTest:    []int{69, 42069},
		BoolTest:        true,
		BoolSliceTest:   []bool{true, false, true, true, false},
		StringTest:      "Chungus - Big Chungus",
		StringSliceTest: []string{"This", "is", "a", "test", "of", "strings"},
		FloatTest:       math.Pi,
		FloatSliceTest:  []float64{math.SqrtPi, math.SqrtE},
	}

	S := test2{
		TestStruct: s,
		Foo: Foo{
			Bar: "rjhwer",
			Baz: 313,
		},
	}

	fmt.Printf("Struct before any encoding/decoding:\n%+v\n", S)

	H, err := DefaultEncode(S)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("Struct to encode:\n%+v\nEncoded Header:\n%+v\n", S, H)

	S2 := test2{}
	if err := DefaultDecode(H, &S2); err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("Pre-encoding struct:\n%+v\nStruct decoded from Header:\n%+v\n", S, S2)
}

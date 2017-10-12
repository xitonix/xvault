package b64

import "fmt"

func ExampleEncode() {
	encoded := Encode([]byte("Plain Text"))
	fmt.Println(string(encoded))
	// Output: UGxhaW4gVGV4dA
}

func ExampleDecode() {
	decoded, err := Decode([]byte{85, 71, 120, 104, 97, 87, 52, 103, 86, 71, 86, 52, 100, 65})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(decoded))
	// Output: Plain Text
}

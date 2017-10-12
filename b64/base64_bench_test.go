package b64

import "testing"

func BenchmarkEncode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Encode([]byte(`Aenean ut rhoncus dolor`))
	}
}

func BenchmarkDecode(b *testing.B) {
	encoded := Encode([]byte(`Aenean ut rhoncus dolor`))
	for i := 0; i < b.N; i++ {
		Decode(encoded)
	}
}

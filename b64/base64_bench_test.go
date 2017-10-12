package b64

import "testing"

func BenchmarkEncodeRawStandard(b *testing.B) {
	encoding := NewRawStandardEncoding()
	for i := 0; i < b.N; i++ {
		encoding.Encode([]byte(`Aenean ut rhoncus dolor`))
	}
}

func BenchmarkDecodeRawStandard(b *testing.B) {
	encoding := NewRawStandardEncoding()
	encoded := encoding.Encode([]byte(`Aenean ut rhoncus dolor`))
	for i := 0; i < b.N; i++ {
		encoding.Decode(encoded)
	}
}

func BenchmarkEncodeRawURL(b *testing.B) {
	encoding := NewRawURLEncoding()
	for i := 0; i < b.N; i++ {
		encoding.Encode([]byte(`Aenean ut rhoncus dolor`))
	}
}

func BenchmarkDecodeRawURL(b *testing.B) {
	encoding := NewRawURLEncoding()
	encoded := encoding.Encode([]byte(`Aenean ut rhoncus dolor`))
	for i := 0; i < b.N; i++ {
		encoding.Decode(encoded)
	}
}

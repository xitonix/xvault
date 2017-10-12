package obfuscate

import "testing"

// Encryption

func BenchmarkEncryptBytes16(b *testing.B) {
	key := make([]byte, 16)
	for i := 0; i < b.N; i++ {
		EncryptBytes(key, []byte("input"))
	}
}

func BenchmarkEncryptBytes24(b *testing.B) {
	key := make([]byte, 24)
	for i := 0; i < b.N; i++ {
		EncryptBytes(key, []byte("input"))
	}
}

func BenchmarkEncryptBytes32(b *testing.B) {
	key := make([]byte, 32)
	for i := 0; i < b.N; i++ {
		EncryptBytes(key, []byte("input"))
	}
}

func BenchmarkEncryptBytesFixed16(b *testing.B) {
	key := make([]byte, 16)
	for i := 0; i < b.N; i++ {
		EncryptBytesFixed(key, []byte("input"))
	}
}

func BenchmarkEncryptBytesFixed24(b *testing.B) {
	key := make([]byte, 24)
	for i := 0; i < b.N; i++ {
		EncryptBytesFixed(key, []byte("input"))
	}
}

func BenchmarkEncryptBytesFixed32(b *testing.B) {
	key := make([]byte, 32)
	for i := 0; i < b.N; i++ {
		EncryptBytesFixed(key, []byte("input"))
	}
}

//Decryption

func BenchmarkDecrypt16(b *testing.B) {
	key := make([]byte, 32)
	encrypted, _ := EncryptBytes(key, []byte("input"))
	for i := 0; i < b.N; i++ {
		DecryptBytes(key, encrypted)
	}
}

func BenchmarkDecrypt24(b *testing.B) {
	key := make([]byte, 24)
	encrypted, _ := EncryptBytes(key, []byte("input"))
	for i := 0; i < b.N; i++ {
		DecryptBytes(key, encrypted)
	}
}

func BenchmarkDecrypt32(b *testing.B) {
	key := make([]byte, 32)
	encrypted, _ := EncryptBytes(key, []byte("input"))
	for i := 0; i < b.N; i++ {
		DecryptBytes(key, encrypted)
	}
}


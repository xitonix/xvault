package hash

import (
	"crypto/sha256"
	"crypto/sha512"
)

// SHA256 returns a 32 bytes SHA256 hash of the input
func SHA256(in []byte) ([]byte, error) {
	h := sha256.New()
	_, err := h.Write([]byte(in))
	if err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

// SHA224 returns a 28 bytes SHA224 hash of the input
func SHA224(in []byte) ([]byte, error) {
	h := sha256.New224()
	_, err := h.Write([]byte(in))
	if err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

// SHA512 returns a 64 bytes SHA512 hash of the input
func SHA512(in []byte) ([]byte, error) {
	h := sha512.New()
	_, err := h.Write([]byte(in))
	if err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

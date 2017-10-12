// Package b64 implements the functionality for encoding and decoding to and from base64 bytes
package b64

import (
	"encoding/base64"
	"io"
)

// Encode encodes the input using base64 raw standard encoder
func Encode(in []byte) []byte {
	buf := make([]byte, base64.RawStdEncoding.EncodedLen(len(in)))
	base64.RawStdEncoding.Encode(buf, in)
	return buf
}

// Decode decodes a base64 encoded []byte using raw standard encoder
func Decode(in []byte) ([]byte, error) {
	buf := make([]byte, base64.RawStdEncoding.DecodedLen(len(in)))
	_, err := base64.RawStdEncoding.Decode(buf, in)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return buf, err
}

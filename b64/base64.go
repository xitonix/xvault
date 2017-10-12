// Package b64 implements the functionality for encoding and decoding to and from base64 bytes
package b64

import (
	"encoding/base64"
	"io"
)

// Base64Encoding encoding interface
type Base64Encoding interface {
	// Encode encodes the input
	Encode(in []byte) []byte
	// Decode decodes a base64 encoded []byte
	Decode(in []byte) ([]byte, error)
}

type encoding struct {
	enc *base64.Encoding
}

// NewRawStandardEncoding creates a new Base64Encoding with internal RawStdEncoding encoding which is the standard raw,
// un-padded base64 encoding, as defined in RFC 4648 section 3.2.
// This is the same as StdEncoding but omits padding characters
func NewRawStandardEncoding() Base64Encoding {
	return &encoding{enc: base64.RawStdEncoding}
}

// NewRawURLEncoding creates a new Base64Encoding with internal RawURLEncoding encoding which is the un-padded alternate base64 encoding defined in RFC 4648.
// It is typically used in URLs and file names. This is the same as URLEncoding but omits padding characters.
func NewRawURLEncoding() Base64Encoding {
	return &encoding{enc: base64.RawURLEncoding}
}

// Encode encodes the input using the specified base64 encoding
func (e *encoding) Encode(in []byte) []byte {
	buf := make([]byte, e.enc.EncodedLen(len(in)))
	e.enc.Encode(buf, in)
	return buf
}

// Decode decodes a base64 encoded []byte using the specified base64 encoding
func (e *encoding) Decode(in []byte) ([]byte, error) {
	buf := make([]byte, e.enc.DecodedLen(len(in)))
	_, err := e.enc.Decode(buf, in)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return buf, err
}

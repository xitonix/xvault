package obfuscate

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"io"
)

// Decoder is the type that decrypts an io Reader into one or more io Writers using the specified master key
type Decoder struct {
	input      io.Reader
	output     io.Writer
	bufferSize int
	master     *MasterKey
}

// NewDecoder creates a new Decoder object
func NewDecoder(bufferSize int, master *MasterKey, input io.Reader, outputs ...io.Writer) *Decoder {
	if bufferSize <= 0 {
		bufferSize = defaultBufferSize
	}

	return &Decoder{
		input:      input,
		output:     io.MultiWriter(outputs...),
		bufferSize: bufferSize,
		master:     master,
	}
}

// Decode decrypts the encoded content of the Reader into the specified Writer(s).
//
// This methods will return an error if the key is invalid or the decryption process fails.
// The content of the input stream must be encoded using the same master key.
func (d *Decoder) Decode() (Status, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	return d.DecodeContext(ctx)
}

// DecodeContext decrypts the encoded content of the Reader into the specified Writer(s) and receives cancellation signal on the context parameter.
//
// It will return an error if the key is invalid or the decryption process fails.
// The content of the input stream must be encoded using the same master key.
func (d *Decoder) DecodeContext(ctx context.Context) (Status, error) {
	if !d.master.isValid() {
		return Failed, errInvalidKey
	}

	cancelled := monitorCancellation(ctx)

	iv, err := d.readMetadata()
	if err != nil {
		return Failed, err
	}

	if *cancelled {
		return Cancelled, nil
	}

	block, err := aes.NewCipher(d.master.key)
	if err != nil {
		return Failed, err
	}

	stream := cipher.NewCFBDecrypter(block, iv)

	return processData(d.input, d.output, d.bufferSize, stream, cancelled)
}

func (d *Decoder) readMetadata() ([]byte, error) {
	meta := make([]byte, signatureLength+aes.BlockSize)
	_, err := d.input.Read(meta)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(d.master.signature, meta[:signatureLength]) {
		return nil, errInvalidSignature
	}

	return meta[signatureLength:], nil
}

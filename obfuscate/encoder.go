package obfuscate

import (
	"io"

	"context"
	"crypto/aes"
	"crypto/cipher"
)

// Encoder is the type that encrypts an io.Reader into one or more io.Writer outputs using the specified master key
type Encoder struct {
	input      io.Reader
	output     io.Writer
	bufferSize int
	master     *MasterKey
}

// NewEncoder creates a new Encoder object
func NewEncoder(bufferSize int, master *MasterKey, input io.Reader, outputs ...io.Writer) *Encoder {
	if bufferSize <= 0 {
		bufferSize = defaultBufferSize
	}

	return &Encoder{
		input:      input,
		output:     io.MultiWriter(outputs...),
		bufferSize: bufferSize,
		master:     master,
	}
}

// Encode encrypts the io.Reader into the specified io.Writer outputs.
// This methods will return an error if the key is invalid or the encryption process fails
func (e *Encoder) Encode() (Status, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	return e.EncodeContext(ctx)
}

// EncodeContext encrypts the io.Reader into the specified io.Writer outputs and receives cancellation signal on the context parameter
// This methods will return an error if the key is invalid or the encryption process fails.
func (e *Encoder) EncodeContext(ctx context.Context) (Status, error) {
	if !e.master.isValid() {
		return Failed, errInvalidKey
	}

	cancelled := monitorCancellation(ctx)

	iv, err := e.writeMetadata()
	if err != nil {
		return Failed, err
	}

	if *cancelled {
		return Cancelled, nil
	}

	block, err := aes.NewCipher(e.master.key)
	if err != nil {
		return Failed, err
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	return processData(e.input, e.output, e.bufferSize, stream, cancelled)
}

func (e *Encoder) writeMetadata() ([]byte, error) {
	iv := getRandomIV()
	_, err := e.output.Write(e.master.signature)
	if err != nil {
		return nil, err
	}

	_, err = e.output.Write(iv)
	if err != nil {
		return nil, err
	}
	return iv, nil
}

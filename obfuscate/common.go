package obfuscate

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"io"

	"github.com/NebulousLabs/fastrand"
	"github.com/xitonix/xvault/b64"
	"github.com/xitonix/xvault/hash"
)

// None represents an empty struct{}
type None struct{}

const (
	defaultBufferSize = 1024
	signatureLength   = 28
	keyLength         = 32
)

var (
	b64Encoding b64.Base64Encoding
)

// SetBase64Encoding sets the base64 encoder used across the package.
// The default is RawStandardEncoding which is the standard raw, un-padded base64 encoding,
// as defined in RFC 4648 section 3.2 and is the same as StdEncoding but omits padding characters.
func SetBase64Encoding(enc b64.Base64Encoding) {
	b64Encoding = enc
}

func init() {
	b64Encoding = b64.NewRawStandardEncoding()
}

func getIV(text []byte, fixed bool) ([]byte, error) {
	if fixed {
		iv, err := hash.SHA224(text)
		if err != nil {
			return nil, err
		}
		return iv[:aes.BlockSize], nil
	}

	return getRandomIV(), nil
}

func getRandomIV() []byte {
	iv := make([]byte, aes.BlockSize)
	fastrand.Read(iv)
	return iv
}

func processData(input io.Reader, output io.Writer, bufferSize int, stream cipher.Stream, cancelled *bool) (Status, error) {
	buffer := make([]byte, bufferSize)
	for {
		if *cancelled {
			break
		}
		count, err := input.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			return Failed, err
		}
		if count > 0 {
			stream.XORKeyStream(buffer[:count], buffer[:count])
			_, err := output.Write(buffer[:count])
			if err != nil && err != io.EOF {
				return Failed, err
			}
		}
	}
	return Completed, nil
}

func monitorCancellation(ctx context.Context) *bool {
	var cancelled bool
	go func() {
		select {
		case <-ctx.Done():
			cancelled = true
			return
		}
	}()
	return &cancelled
}

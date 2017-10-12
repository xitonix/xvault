package obfuscate

import (
	"bytes"
	"strings"
	"unicode/utf8"
	"github.com/xitonix/xvault/b64"
	"github.com/xitonix/xvault/hash"
)

// MasterKey cryptography master key
type MasterKey struct {
	// key 32 bytes cryptography key
	key []byte
	// signature 28 bytes obfuscation signature
	signature []byte
	// password encrypted password which can be store in the password file
	password []byte
}

// Signature returns the 28 bytes obfuscation signature
func (k *MasterKey) Signature() []byte {
	return k.signature
}

// Password returns the encrypted password.
// It's safe to store the encrypted password into a file.
func (k *MasterKey) Password() []byte {
	return k.password
}

// Validate returns true if the same password has been used to generate the master key
func (k *MasterKey) Validate(pass string) bool {
	defer func() {
		recover()
	}()

	if !k.isValid() {
		return false
	}

	key, err := KeyFromPassword(pass)
	if err != nil || !k.isValid() {
		return false
	}

	return bytes.Equal(k.password, key.password)
}

// KeyFromPassword creates a cryptography master key based on the provided password
func KeyFromPassword(pass string) (*MasterKey, error) {
	if len(strings.TrimSpace(pass)) == 0 {
		return nil, errEmptyPassword
	}

	if utf8.RuneCount([]byte(pass)) < 8 {
		return nil, errInvalidPassword
	}

	b := promotePassword(pass)

	signature, err := hash.SHA224(b)
	if err != nil {
		return nil, err
	}

	base64 := b64.Encode(b)

	key, err := hash.SHA256(base64)
	if err != nil {
		return nil, err
	}

	// We need to use the same IV to encrypt the password
	// Otherwise the encryption result of two same passwords will be different
	encrypted, err := EncryptBytesFixed(key, b)

	if err != nil {
		return nil, err
	}

	hashedSignature, err := hash.SHA512(signature)
	if err != nil {
		return nil, err
	}
	password := append(hashedSignature, encrypted...)
	password = b64.Encode(password)

	return &MasterKey{
		key:       key,
		signature: signature,
		password:  password,
	}, nil
}

func (k *MasterKey) isValid() bool {
	return k != nil &&
		len(k.key) == keyLength &&
		len(k.signature) == signatureLength &&
		len(k.password) > 0
}

func promotePassword(pass string) []byte {
	return b64.Encode([]byte(strings.ToLower(pass) + pass + strings.ToUpper(pass)))
}

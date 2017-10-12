package obfuscate

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"github.com/xitonix/xvault/b64"
)

// EncryptBytesFixed encrypts a string using a random crypto IV
// Use this method when you want the encryption result of two same input to be be different
func EncryptBytes(key, text []byte) ([]byte, error) {
	cipherText, base64, err := allocate(text, false)
	if err != nil {
		return nil, err
	}
	return encrypt(cipherText, base64, key)
}

// EncryptBytesFixed encrypts a string using a same crypto IV
// Use this method when you want the encryption result of two same input to be be the same
func EncryptBytesFixed(key, text []byte) ([]byte, error) {
	cipherText, base64, err := allocate(text, true)
	if err != nil {
		return nil, err
	}

	return encrypt(cipherText, base64, key)
}

// DecryptBytes decrypts a string
func DecryptBytes(key, textBytes []byte) ([]byte, error) {
	if len(textBytes) < aes.BlockSize {
		return nil, errors.New("invalid encrypted bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	iv := textBytes[:aes.BlockSize]
	textBytes = textBytes[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(textBytes, textBytes)
	data, err := b64.Decode(textBytes)

	if err != nil {
		return nil, err
	}
	return data, nil
}

func allocate(text []byte, fixed bool) ([]byte, []byte, error) {
	b := b64.Encode(text)

	cipherText := make([]byte, aes.BlockSize+len(b))
	iv, err := getIV(text, fixed)
	if err != nil {
		return nil, nil, err
	}
	copy(cipherText[:aes.BlockSize], iv)
	return cipherText, b, nil
}

func encrypt(cipherText, base64, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	cfb := cipher.NewCFBEncrypter(block, cipherText[:aes.BlockSize])
	cfb.XORKeyStream(cipherText[aes.BlockSize:], base64)
	return cipherText, nil
}

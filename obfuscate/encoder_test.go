package obfuscate

import (
	"testing"

	"github.com/mattetti/filebuffer"
	"github.com/xitonix/xvault/assert"
)

func TestEncode(t *testing.T) {
	const (
		ivLength        = 16
		signatureLength = 28
	)
	testCases := []struct {
		title                    string
		expectedLength           int64
		input                    string
		bufferSize               int
		expectedBufferSize       int
		expectKeyGenerationError bool
		password                 string
	}{
		{
			title:              "empty_input",
			expectedLength:     ivLength + signatureLength,
			input:              "",
			bufferSize:         100,
			expectedBufferSize: 100,
			password:           "password",
		},
		{
			title:              "whitespace_input",
			expectedLength:     ivLength + signatureLength + 1,
			input:              " ",
			bufferSize:         100,
			expectedBufferSize: 100,
			password:           "password",
		},
		{
			title:              "non_empty_input",
			expectedLength:     ivLength + signatureLength + 2,
			input:              "Go",
			bufferSize:         100,
			expectedBufferSize: 100,
			password:           "password",
		},
		{
			title:              "invalid_buffer_size_should_get_fixed_automatically",
			expectedLength:     ivLength + signatureLength + 2,
			input:              "Go",
			bufferSize:         0,
			expectedBufferSize: defaultBufferSize,
			password:           "password",
		},
		{
			title:                    "short_password_must_fail_to_generate_the_key",
			password:                 "short",
			expectKeyGenerationError: true,
		},
		{
			title:                    "empty_password_must_fail_to_generate_the_key",
			password:                 "",
			expectKeyGenerationError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			in := filebuffer.New([]byte(tc.input))
			out := filebuffer.New(nil)
			key, err := KeyFromPassword(tc.password)
			if !assert.Errors(t, tc.expectKeyGenerationError, err, assert.Fields{"password": tc.password}) {
				return
			}
			encoder := NewEncoder(tc.bufferSize, key, in, out)
			status, err := encoder.Encode()
			if err != nil {
				t.Errorf("failed to encode: %v", err)
				return
			}

			if status != Completed {
				t.Errorf("expected encoding status to be '%s', actual '%s'", Completed, status)
			}

			if tc.expectedLength != out.Index {
				t.Errorf("encrypted output length expected to be %d, but it was %d", tc.expectedLength, out.Index)
			}

			if encoder.bufferSize != tc.expectedBufferSize {
				t.Errorf("expected buffer size %d, but got %d", tc.expectedBufferSize, encoder.bufferSize)
			}
		})
	}
}

func TestEncodeMultipleOutputs(t *testing.T) {
	const (
		ivLength        = 16
		signatureLength = 28
	)
	testCases := []struct {
		title              string
		expectedLength     int64
		input              string
		bufferSize         int
		expectedBufferSize int
		password           string
	}{
		{
			title:              "empty_input",
			expectedLength:     ivLength + signatureLength,
			input:              "",
			bufferSize:         100,
			expectedBufferSize: 100,
			password:           "password",
		},
		{
			title:              "whitespace_input",
			expectedLength:     ivLength + signatureLength + 1,
			input:              " ",
			bufferSize:         100,
			expectedBufferSize: 100,
			password:           "password",
		},
		{
			title:              "non_empty_input",
			expectedLength:     ivLength + signatureLength + 2,
			input:              "Go",
			bufferSize:         100,
			expectedBufferSize: 100,
			password:           "password",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			in := filebuffer.New([]byte(tc.input))
			out1 := filebuffer.New(nil)
			out2 := filebuffer.New(nil)
			key, err := KeyFromPassword(tc.password)
			if !assert.Errors(t, false, err, assert.Fields{"password": tc.password}) {
				return
			}
			encoder := NewEncoder(tc.bufferSize, key, in, out1, out2)
			status, err := encoder.Encode()
			if err != nil {
				t.Errorf("failed to encode: %v", err)
				return
			}

			if status != Completed {
				t.Errorf("expected encoding status to be '%s', actual '%s'", Completed, status)
			}

			if tc.expectedLength != out1.Index {
				t.Errorf("encrypted output#1 length expected to be %d, but it was %d", tc.expectedLength, out1.Index)
			}

			if tc.expectedLength != out2.Index {
				t.Errorf("encrypted output#2 length expected to be %d, but it was %d", tc.expectedLength, out2.Index)
			}

			if encoder.bufferSize != tc.expectedBufferSize {
				t.Errorf("expected buffer size %d, but got %d", tc.expectedBufferSize, encoder.bufferSize)
			}
		})
	}
}

func TestEncodeInvalidKey(t *testing.T) {
	testCases := []struct {
		title  string
		master *MasterKey
	}{
		{
			title: "key_with_invalid_signature_length",
			master: &MasterKey{
				signature: make([]byte, 6),
				key:       make([]byte, keyLength),
				password:  make([]byte, 10),
			},
		},
		{
			title: "key_with_invalid_key_length",
			master: &MasterKey{
				signature: make([]byte, signatureLength),
				key:       make([]byte, 7),
				password:  make([]byte, 10),
			},
		},
		{
			title: "key_with_invalid_password_length",
			master: &MasterKey{
				signature: make([]byte, signatureLength),
				key:       make([]byte, keyLength),
				password:  []byte{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			in := filebuffer.New([]byte("input"))
			out := filebuffer.New(nil)
			encoder := NewEncoder(defaultBufferSize, tc.master, in, out)
			status, err := encoder.Encode()
			assert.Errors(t, true, err, nil)
			if status != Failed {
				t.Errorf("expected encoding status to be '%s', actual %s", Failed, status)
			}
		})
	}
}

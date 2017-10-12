package obfuscate

import (
	"testing"

	"github.com/mattetti/filebuffer"
	"github.com/xitonix/xvault/assert"
)

func TestDecode(t *testing.T) {
	const (
		ivLength = 16
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

			decodedAndAssert(t, out.Buff.Bytes(), key, tc.input)
		})
	}
}

func TestDecodeMultipleEncodedOutputs(t *testing.T) {
	const (
		ivLength = 16
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
		{
			title:              "invalid_buffer_size_should_get_fixed_automatically",
			expectedLength:     ivLength + signatureLength + 2,
			input:              "Go",
			bufferSize:         0,
			expectedBufferSize: defaultBufferSize,
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

			decodedAndAssert(t, out1.Buff.Bytes(), key, tc.input)
			decodedAndAssert(t, out2.Buff.Bytes(), key, tc.input)
		})
	}
}

func TestDecodeMultipleDecodedOutputs(t *testing.T) {
	const (
		ivLength = 16
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
		{
			title:              "invalid_buffer_size_should_get_fixed_automatically",
			expectedLength:     ivLength + signatureLength + 2,
			input:              "Go",
			bufferSize:         0,
			expectedBufferSize: defaultBufferSize,
			password:           "password",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			in := filebuffer.New([]byte(tc.input))
			out := filebuffer.New(nil)
			key, err := KeyFromPassword(tc.password)
			if !assert.Errors(t, false, err, assert.Fields{"password": tc.password}) {
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

			in = filebuffer.New(out.Buff.Bytes())
			out1 := filebuffer.New(nil)
			out2 := filebuffer.New(nil)

			decoder := NewDecoder(defaultBufferSize, key, in, out1, out2)
			status, err = decoder.Decode()
			if err != nil {
				t.Errorf("failed to decode: %v", err)
				return
			}

			if status != Completed {
				t.Errorf("expected decoding status to be '%s', actual '%s'", Completed, status)
			}

			actual := string(out1.Buff.Bytes())
			if actual != tc.input {
				t.Errorf("expected %s, received %s", tc.input, actual)
			}

			actual = string(out2.Buff.Bytes())
			if actual != tc.input {
				t.Errorf("expected %s, received %s", tc.input, actual)
			}
		})
	}
}

func TestDecodeInvalidKey(t *testing.T) {
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
			decoder := NewDecoder(defaultBufferSize, tc.master, in, out)
			status, err := decoder.Decode()
			assert.Errors(t, true, err, nil)
			if status != Failed {
				t.Errorf("expected decoding status to be '%s', actual '%s'", Failed, status)
			}
		})
	}
}

func decodedAndAssert(t *testing.T, encoded []byte, master *MasterKey, expected string) {
	t.Helper()
	in := filebuffer.New(encoded)
	out := filebuffer.New(nil)

	decoder := NewDecoder(defaultBufferSize, master, in, out)
	status, err := decoder.Decode()
	if err != nil {
		t.Errorf("failed to decode: %v", err)
		return
	}

	if status != Completed {
		t.Errorf("expected decoding status to be '%s', actual '%s'", Completed, status)
	}

	actual := string(out.Buff.Bytes())
	if actual != expected {
		t.Errorf("expected %s, received %s", expected, actual)
	}
}

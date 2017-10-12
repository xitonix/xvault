package obfuscate

import (
	"bytes"
	"testing"
	"github.com/xitonix/xvault/assert"
)

func TestEncryptBytesFixed(t *testing.T) {
	testCases := []struct {
		title       string
		expectError bool
		input       string
		keys        [][]byte
	}{
		{
			title: "valid_keys_must_produce_valid_encrypted_output",
			input: "a",
			keys: [][]byte{
				make([]byte, 16),
				make([]byte, 24),
				make([]byte, 32),
			},
		},
		{
			title:       "invalid_keys_must_fail_to_encrypt",
			input:       "a",
			expectError: true,
			keys: [][]byte{
				make([]byte, 7),
				{},
			},
		},
		{
			title: "empty_input_is_valid",
			input: "",
			keys: [][]byte{
				make([]byte, 16),
				make([]byte, 24),
				make([]byte, 32),
			},
		},
		{
			title: "whitespace_input_is_valid",
			input: "   ",
			keys: [][]byte{
				make([]byte, 16),
				make([]byte, 24),
				make([]byte, 32),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			for i, key := range tc.keys {
				fields := assert.Fields{"key index": i, "input": tc.input}

				actual, err := EncryptBytesFixed(key, []byte(tc.input))
				if !assert.Errors(t, tc.expectError, err, fields) {
					continue
				}

				if len(actual) == 0 {
					t.Errorf("encrypted string was empty (key index: %d)", i)
					continue
				}

				text, err := DecryptBytes(key, actual)

				if !assert.Errors(t, tc.expectError, err, fields) {
					continue
				}

				if string(text) != tc.input {
					t.Errorf("Decryption failed. Expected %s, but received %v (key index: %d)", tc.input, string(text), i)
					continue
				}
			}
		})
	}
}

func TestEncryptBytesFixedOutputComparison(t *testing.T) {
	testCases := []struct {
		title          string
		input1, input2 string
		keys           [][]byte
	}{
		{
			title:  "same_input_must_produce_same_encrypted_result",
			input1: "a",
			input2: "a",
			keys: [][]byte{
				make([]byte, 16),
				make([]byte, 24),
				make([]byte, 32),
			},
		},
		{
			title:  "different_input_must_produce_different_encrypted_result",
			input1: "a",
			input2: "b",
			keys: [][]byte{
				make([]byte, 16),
				make([]byte, 24),
				make([]byte, 32),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			for i, key := range tc.keys {
				fields := assert.Fields{
					"key index": i,
					"input1":    tc.input1,
					"input2":    tc.input2,
				}
				result1, err := EncryptBytesFixed(key, []byte(tc.input1))
				if !assert.Errors(t, false, err, fields) {
					continue
				}

				if len(result1) == 0 {
					t.Errorf("encrypted string was empty (%s)", fields.String())
					continue
				}

				result2, err := EncryptBytesFixed(key, []byte(tc.input2))
				if !assert.Errors(t, false, err, fields) {
					continue
				}

				if len(result2) == 0 {
					t.Errorf("encrypted string was empty (%s)", fields.String())
					continue
				}

				if tc.input1 == tc.input2 && !bytes.Equal(result1, result2) {
					t.Errorf("encrypted strings are not equal (%s)", fields.String())
					continue
				}

				if tc.input1 != tc.input2 && bytes.Equal(result1, result2) {
					t.Errorf("encryption results should not be equal (%s)", fields.String())
				}
			}
		})
	}
}

func TestEncryptBytes(t *testing.T) {
	testCases := []struct {
		title       string
		expectError bool
		input       string
		keys        [][]byte
	}{
		{
			title: "valid_keys_must_produce_valid_encrypted_output",
			input: "a",
			keys: [][]byte{
				make([]byte, 16),
				make([]byte, 24),
				make([]byte, 32),
			},
		},
		{
			title:       "invalid_keys_must_fail_to_encrypt",
			input:       "a",
			expectError: true,
			keys: [][]byte{
				make([]byte, 7),
				{},
			},
		},
		{
			title: "empty_input_is_valid",
			input: "",
			keys: [][]byte{
				make([]byte, 16),
				make([]byte, 24),
				make([]byte, 32),
			},
		},
		{
			title: "whitespace_input_is_valid",
			input: "   ",
			keys: [][]byte{
				make([]byte, 16),
				make([]byte, 24),
				make([]byte, 32),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			for i, key := range tc.keys {
				fields := assert.Fields{"key index": i, "input": tc.input}

				actual, err := EncryptBytes(key, []byte(tc.input))
				if !assert.Errors(t, tc.expectError, err, fields) {
					continue
				}

				if len(actual) == 0 {
					t.Errorf("encrypted string was empty (key index: %d)", i)
					continue
				}

				text, err := DecryptBytes(key, actual)

				if !assert.Errors(t, tc.expectError, err, fields) {
					continue
				}

				if string(text) != tc.input {
					t.Errorf("Decryption failed. Expected %s, but received %v (key index: %d)", tc.input, string(text), i)
					continue
				}
			}
		})
	}
}

func TestEncryptBytesOutputComparison(t *testing.T) {
	testCases := []struct {
		title          string
		input1, input2 string
		keys           [][]byte
	}{
		{
			title:  "same_input_must_produce_different_encrypted_result",
			input1: "a",
			input2: "a",
			keys: [][]byte{
				make([]byte, 16),
				make([]byte, 24),
				make([]byte, 32),
			},
		},
		{
			title:  "different_input_must_produce_different_encrypted_result",
			input1: "a",
			input2: "b",
			keys: [][]byte{
				make([]byte, 16),
				make([]byte, 24),
				make([]byte, 32),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			for i, key := range tc.keys {
				fields := assert.Fields{
					"key index": i,
					"input1":    tc.input1,
					"input2":    tc.input2,
				}
				result1, err := EncryptBytes(key, []byte(tc.input1))
				if !assert.Errors(t, false, err, fields) {
					continue
				}

				if len(result1) == 0 {
					t.Errorf("encrypted string was empty (%s)", fields.String())
					continue
				}

				result2, err := EncryptBytes(key, []byte(tc.input2))
				if !assert.Errors(t, false, err, fields) {
					continue
				}

				if len(result2) == 0 {
					t.Errorf("encrypted string was empty (%s)", fields.String())
					continue
				}

				if bytes.Equal(result1, result2) {
					t.Errorf("encryption results should be different (%s)", fields.String())
				}
			}
		})
	}
}

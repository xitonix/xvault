package b64

import (
	"github.com/xitonix/xvault/assert"
	"testing"
)

func TestBase64EncodingDecoding(t *testing.T) {
	testCases := []struct {
		title string
		input string
	}{
		{
			title: "empty_plain_text_is_valid",
			input: "",
		},
		{
			title: "whitespace_text_is_valid",
			input: "  ",
		},
		{
			title: "short_input",
			input: "a",
		},
		{
			title: "long_input",
			input: `Aenean ut rhoncus dolor,
			et porttitor dui. Donec a orci in justo maximus interdum.
			Phasellus semper, nisl ac semper dictum, risus orci facilisis
			lorem, nec interdum diam risus a libero. Aenean id fermentum mauris.
			Nunc consequat finibus tortor, nec feugiat justo consectetur sed.
			Proin non tincidunt odio, ac imperdiet tellus.
			Vestibulum nec quam vitae erat tincidunt dapibus eget ut sapien.
			Nunc lacinia arcu eros, id laoreet nisl fermentum eu.
			Etiam venenatis ligula felis, et ultrices neque ultrices id.
			Nam molestie ultrices nisl sit amet consectetur.
			Proin quis auctor ante. Aliquam erat volutpat`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			encoded := Encode([]byte(tc.input))

			if len(tc.input) != 0 && len(encoded) == 0 {
				t.Error("encoded bytes is empty")
			}

			decoded, err := Decode(encoded)
			if !assert.Errors(t, false, err, nil) {
				return
			}

			if string(decoded) != tc.input {
				t.Errorf("wrong decoded result. expected '%v', actual '%v", tc.input, string(decoded))
			}
		})
	}
}

func TestBase64DecodeInvalidInput(t *testing.T) {
	testCases := []struct {
		title       string
		input       string
		expectError bool
	}{
		{
			title: "empty_plain_text_is_valid",
			input: "",
		},
		{
			title:       "whitespace_text_is_not_valid",
			input:       "  ",
			expectError: true,
		},
		{
			title:       "invalid_base64_input",
			input:       "Aenean ut rhoncus dolor",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			result, err := Decode([]byte(tc.input))
			assert.Errors(t, tc.expectError, err, nil)
			if len(result) > 0 {
				t.Errorf("the result was supposed to be empty, but it was %v", result)
			}
		})
	}
}

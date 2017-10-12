package obfuscate

import (
	"testing"
)

func TestMasterFromPassword(t *testing.T) {
	testCases := []struct {
		title         string
		expectedError error
		password      string
	}{
		{
			title:         "empty_password_is_not_valid",
			password:      "",
			expectedError: errEmptyPassword,
		},
		{
			title:         "whitespace_password_is_not_valid",
			password:      "    ",
			expectedError: errEmptyPassword,
		},
		{
			title:         "passwords_shorter_than_eight_characters_are_not_valid",
			password:      "1234567",
			expectedError: errInvalidPassword,
		},
		{
			title:    "passwords_with_at_least_eight_characters_are_valid",
			password: "12345678",
		},
		{
			title:    "passwords_with_more_than_eight_characters_are_valid",
			password: "1234567891012345678910",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			key, err := KeyFromPassword(tc.password)
			if err != nil {
				if tc.expectedError != err {
					t.Errorf("expected '%v' error, actual '%v'", tc.expectedError, err)
				}
				return
			}
			if !key.isValid() {
				t.Errorf("Invalid master key: %+v", key)
			}
		})
	}

}

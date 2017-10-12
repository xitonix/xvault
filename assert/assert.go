// Package assert includes some helper methods used for testing
package assert

import (
	"testing"
)

// Errors checks the validity of the expected error and returns false if the assertion failed
func Errors(t *testing.T, expectError bool, err error, fields Fields) bool {
	t.Helper()

	if expectError && err == nil {
		t.Errorf("Expected an error, but received 'nil' (%s)", fields.String())
	}

	if !expectError && err != nil {
		t.Errorf("No error was expected, but received '%v' (%s)", err, fields.String())
	}

	return !expectError
}

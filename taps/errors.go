package taps

import "errors"

var (
	// ErrInvalidDirectory raised if the specified path is not a valid path to a directory
	ErrInvalidDirectory = errors.New("the specified path is not a directory")
)

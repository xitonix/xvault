package obfuscate

import "errors"

var (
	errInvalidSignature = errors.New("invalid file content")
	errInvalidKey       = errors.New("invalid key")
	errEmptyPassword    = errors.New("password cannot be empty")
	errInvalidPassword  = errors.New("password must be at least eight characters long")

	// ErrOperationInProgress is the result of any invalid operation on an entity which is already being processed
	ErrOperationInProgress = errors.New("the operation is in progress")
	// ErrClosedTap will be raised if the user tries to push to the pipe from a closed Tap
	ErrClosedTap = errors.New("cannot push from a closed tap")
)

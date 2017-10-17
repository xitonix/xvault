package obfuscate

import "errors"

var (
	errInvalidSignature = errors.New("invalid signature")
	errInvalidKey       = errors.New("invalid key")
	errEmptyPassword    = errors.New("password cannot be empty")
	errInvalidPassword  = errors.New("password must be at least eight characters long")
	// ErrOperationInProgress an invalid request has been sent to an in-progress operation
	ErrOperationInProgress = errors.New("the operation is in progress")
)

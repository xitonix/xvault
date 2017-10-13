package obfuscate

// Result represents the progress details of a Task
type Result struct {
	// Status the status of the operation
	Status Status

	// Error the error details of a failed Task
	Error error

	Metadata map[string]interface{}
}

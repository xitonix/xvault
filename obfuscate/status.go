package obfuscate

// Status operation status
type Status int8

const (
	// Queued indicates that the operation has been queued
	Queued Status = iota
	// Completed indicates that the operation has been completed successfully
	Completed
	// Cancelled indicates that the operation has been cancelled by the user
	Cancelled
	// Failed indicates that the operation has been failed
	Failed
)

// String returns the string representation of the operation status
func (s Status) String() string {
	switch s {
	case Queued:
		return "queued"
	case Completed:
		return "completed"
	case Cancelled:
		return "cancelled"
	case Failed:
		return "failed"
	}
	return "unknown"
}

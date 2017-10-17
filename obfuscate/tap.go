package obfuscate

// Tap is the interface for the types responsible to send work units to an Engine
type Tap interface {
	// Open opens the tap and starts pushing work units into the associated stream.
	// The engine will automatically open the tap when it starts, so you SHOULD NOT call this method explicitly.
	// Make sure that the implementation of this function IS NOT blocking.
	Open()
	// Close closes the tap and stops pushing work units into the associated stream.
	// The engine will automatically shutdown the tap when it's been stopped, so there is no need for you to explicitly call this method.
	// Make sure the implementation of this method blocks until all the tap's internal resources are released
	Close()
	// IsOpen returns true if the tap is open
	IsOpen() bool
	// Pipe returns the work list pipe attached to the tap
	Pipe() WorkList
}

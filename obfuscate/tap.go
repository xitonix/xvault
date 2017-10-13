package obfuscate

// Tap is the interface for the types responsible to send work units to an instance of Work Bucket
type Tap interface {
	// Open opens a tap and starts pushing work units into the associated stream.
	// The work bucket will automatically open the taps, so there is no need for you to explicitly call this method.
	// NOTE: The implementation of this function SHOULD NOT be blocking.
	Open()
	// Close closes the tap and stops pushing work units into the associated stream.
	// The work bucket will automatically shutdown the taps, so there is no need for you to explicitly call this method.
	// NOTE: Make sure the implementation of this method blocks until all the tap's internal resources are released
	Close()
	// IsOpen returns true if the tap is open
	IsOpen() bool

	Requests() RequestChannel
}

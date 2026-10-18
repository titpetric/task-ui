package runner

// Event is what a subscriber receives: terminal output, or the status the
// run ended with. A status event is the last one.
type Event struct {
	Output []byte
	Status *Status
}

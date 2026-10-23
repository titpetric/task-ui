package runner

// Event is what a subscriber receives: terminal output, the title the
// terminal set, "" when it cleared it, or the status the run ended
// with. A status event is the last one.
type Event struct {
	Output []byte
	Title  *string
	Status *Status
}

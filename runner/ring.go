package runner

// ring keeps the last size bytes written to it, so a window attaching
// late sees what it missed without the run holding everything.
type ring struct {
	buf  []byte
	size int
}

func newRing(size int) *ring {
	return &ring{size: size}
}

// Write keeps the tail of everything written so far.
func (r *ring) Write(p []byte) (int, error) {
	r.buf = append(r.buf, p...)
	if len(r.buf) > r.size {
		r.buf = append([]byte{}, r.buf[len(r.buf)-r.size:]...)
	}
	return len(p), nil
}

// Bytes returns a copy of the tail the ring keeps.
func (r *ring) Bytes() []byte {
	return append([]byte{}, r.buf...)
}

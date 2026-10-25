package storage

import (
	"time"
)

// Thread is a prompt thread as the data retention rules see it: the
// first note, when its last reply came, the first note's own time
// without one, and whether any of its notes is an open or working
// instruction.
type Thread struct {
	ID     string
	LastAt *time.Time
	Open   bool
}

// Package run is the execution context: an active run of a job, the
// registry that allows one run per job across the instance, and the
// process a run drives on a pseudo terminal.
package run

import (
	"sync"
	"time"
)

// Run is an active run of a job. The registry holds one per key, which is
// what keeps two copies of the same job in a folder from running at the
// same time, while different jobs run side by side.
type Run struct {
	ID      string    `json:"id"`
	Key     string    `json:"key"`
	Folder  string    `json:"folder"`
	Job     string    `json:"job"`
	User    string    `json:"user"`
	Started time.Time `json:"started"`
	Forced  bool      `json:"forced,omitempty"`

	mu      sync.Mutex
	killed  bool
	stop    func()
	done    chan struct{}
	release sync.Once
}

// Kill asks the run to stop. It is safe to call before the process starts
// and more than once.
func (r *Run) Kill() {
	r.mu.Lock()
	r.killed = true
	stop := r.stop
	r.mu.Unlock()
	if stop != nil {
		stop()
	}
}

// Killed reports whether Kill was called.
func (r *Run) Killed() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.killed
}

// SetStop installs the function that stops the process. A run killed before
// the process started is stopped immediately.
func (r *Run) SetStop(stop func()) {
	r.mu.Lock()
	r.stop = stop
	killed := r.killed
	r.mu.Unlock()
	if killed {
		stop()
	}
}

// Done is closed when the run is released.
func (r *Run) Done() <-chan struct{} {
	return r.done
}

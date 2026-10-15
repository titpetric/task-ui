package internal

import (
	"errors"
	"sort"
	"sync"
	"time"
)

// ErrBusy is returned when a runbook is already running.
var ErrBusy = errors.New("already running")

// Run is an active run of a runbook. The registry holds one per key, which
// is what keeps two copies of the same runbook in a project from running at
// the same time, while different runbooks run side by side.
type Run struct {
	ID      string    `json:"id"`
	Key     string    `json:"key"`
	Project string    `json:"project"`
	Runbook string    `json:"runbook"`
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

// Runs is the instance wide registry of active runs.
type Runs struct {
	mu     sync.Mutex
	active map[string]*Run

	// KillTimeout bounds how long a forced run waits for the previous run
	// to let go.
	KillTimeout time.Duration
}

// NewRuns returns an empty registry.
func NewRuns() *Runs {
	return &Runs{
		active:      map[string]*Run{},
		KillTimeout: 15 * time.Second,
	}
}

// Acquire registers a run under its key. When the key is taken it returns
// the active run and ErrBusy, unless force is set: then the active run is
// killed, and Acquire waits for it to be released before taking its place.
func (rs *Runs) Acquire(run *Run, force bool) (*Run, error) {
	deadline := time.Now().Add(rs.KillTimeout)
	for {
		rs.mu.Lock()
		current, busy := rs.active[run.Key]
		if !busy {
			run.done = make(chan struct{})
			run.Forced = force
			rs.active[run.Key] = run
			rs.mu.Unlock()
			return run, nil
		}
		rs.mu.Unlock()

		if !force {
			return current, ErrBusy
		}

		current.Kill()
		select {
		case <-current.done:
		case <-time.After(time.Until(deadline)):
			return current, ErrBusy
		}
	}
}

// Release removes a run from the registry.
func (rs *Runs) Release(run *Run) {
	run.release.Do(func() {
		rs.mu.Lock()
		if rs.active[run.Key] == run {
			delete(rs.active, run.Key)
		}
		rs.mu.Unlock()
		close(run.done)
	})
}

// Get returns the active run for a key.
func (rs *Runs) Get(key string) (*Run, bool) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	run, ok := rs.active[key]
	return run, ok
}

// List returns the active runs, oldest first.
func (rs *Runs) List() []*Run {
	rs.mu.Lock()
	result := make([]*Run, 0, len(rs.active))
	for _, run := range rs.active {
		result = append(result, run)
	}
	rs.mu.Unlock()
	sort.Slice(result, func(i, j int) bool { return result[i].Started.Before(result[j].Started) })
	return result
}

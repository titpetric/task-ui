package run

import (
	"errors"
	"sort"
	"sync"
	"time"
)

// ErrBusy is returned when a job is already running.
var ErrBusy = errors.New("already running")

// Registry is the instance wide registry of active runs.
type Registry struct {
	mu     sync.Mutex
	active map[string]*Run

	// KillTimeout bounds how long a forced run waits for the previous run
	// to let go.
	KillTimeout time.Duration
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{
		active:      map[string]*Run{},
		KillTimeout: 15 * time.Second,
	}
}

// Acquire registers a run under its key. When the key is taken it returns
// the active run and ErrBusy, unless force is set: then the active run is
// killed, and Acquire waits for it to be released before taking its place.
func (rs *Registry) Acquire(run *Run, force bool) (*Run, error) {
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
func (rs *Registry) Release(run *Run) {
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
func (rs *Registry) Get(key string) (*Run, bool) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	run, ok := rs.active[key]
	return run, ok
}

// List returns the active runs, oldest first.
func (rs *Registry) List() []*Run {
	rs.mu.Lock()
	result := make([]*Run, 0, len(rs.active))
	for _, run := range rs.active {
		result = append(result, run)
	}
	rs.mu.Unlock()
	sort.Slice(result, func(i, j int) bool { return result[i].Started.Before(result[j].Started) })
	return result
}

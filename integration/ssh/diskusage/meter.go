package diskusage

import (
	"sync"
	"time"
)

// Meter measures directories and keeps each result for TTL, so a page
// that shows a large checkout does not walk it on every view.
type Meter struct {
	TTL time.Duration

	mu    sync.Mutex
	cache map[string]measured
}

type measured struct {
	usage Usage
	at    time.Time
}

// Measure returns the usage of a directory, measured at most TTL ago.
func (m *Meter) Measure(dir string) Usage {
	m.mu.Lock()
	if hit, ok := m.cache[dir]; ok && time.Since(hit.at) < m.TTL {
		m.mu.Unlock()
		return hit.usage
	}
	m.mu.Unlock()

	usage := Measure(dir)

	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cache == nil {
		m.cache = map[string]measured{}
	}
	m.cache[dir] = measured{usage: usage, at: time.Now()}
	return usage
}

// Forget drops what the meter keeps for a directory, after a change to it.
func (m *Meter) Forget(dir string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.cache, dir)
}

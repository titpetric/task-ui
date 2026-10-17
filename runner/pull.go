package runner

import (
	"context"
	"fmt"
	"sync"
)

// pull brings a folder that is a checkout up to date, in the run's own
// terminal. Runs of one folder pull one at a time: two pulls of the same
// checkout at once fight over the index lock, and Run all starts them
// together.
func (m *Manager) pull(ctx context.Context, session *Session) error {
	f := session.Folder
	if m.opts.Git == nil || !m.opts.Git.IsRepo(f.Dir) {
		return nil
	}
	lock := m.pullLock(f.Dir)
	lock.Lock()
	defer lock.Unlock()

	fmt.Fprint(session, "\x1b[2m$ git pull --ff-only\x1b[0m\r\n")
	if err := m.opts.Git.Pull(ctx, f.Dir, crlfWriter{session}); err != nil {
		return fmt.Errorf("git pull failed: %w", err)
	}
	fmt.Fprint(session, "\r\n")
	return nil
}

// pullLock returns the lock of a checkout.
func (m *Manager) pullLock(dir string) *sync.Mutex {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.pulls == nil {
		m.pulls = map[string]*sync.Mutex{}
	}
	lock, ok := m.pulls[dir]
	if !ok {
		lock = &sync.Mutex{}
		m.pulls[dir] = lock
	}
	return lock
}

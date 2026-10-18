package runner

import (
	"context"
	"fmt"
	"sync"
)

// checkout returns the checkout the run's folder belongs to: the folder
// itself, or the nearest parent inside the root that is one. "" for none.
func (m *Manager) checkout(session *Session) string {
	if m.opts.Git == nil {
		return ""
	}
	return m.opts.Git.Checkout(session.Folder.Dir, session.Folder.Depth())
}

// pull brings the checkout of the run's folder up to date, in the run's
// own terminal; a folder inside a checkout pulls the checkout. Runs of one
// checkout pull one at a time: two pulls at once fight over the index
// lock, and Run all starts them together.
func (m *Manager) pull(ctx context.Context, session *Session) error {
	dir := m.checkout(session)
	if dir == "" {
		return nil
	}
	lock := m.pullLock(dir)
	lock.Lock()
	defer lock.Unlock()

	fmt.Fprint(session, "\x1b[2m$ git pull --ff-only\x1b[0m\r\n")
	if err := m.opts.Git.Pull(ctx, dir, crlfWriter{session}); err != nil {
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

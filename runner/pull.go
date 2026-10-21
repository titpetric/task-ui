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

// pull brings the checkouts of the run up to date, in the run's own
// terminal: the workspace's first, so its tooling is current, then the
// folder's when that differs; a folder inside a checkout pulls the
// checkout. Runs of one checkout pull one at a time: two pulls at once
// fight over the index lock, and a workflow starts them one after another.
func (m *Manager) pull(ctx context.Context, session *Session) error {
	if m.opts.Git == nil {
		return nil
	}
	ws := session.workspace
	if ws != "" && m.opts.Git.IsRepo(ws) {
		if err := m.pullDir(ctx, session, ws, "git -C "+ws+" pull --ff-only"); err != nil {
			return err
		}
	}
	dir := m.checkout(session)
	if dir == "" || dir == ws {
		return nil
	}
	return m.pullDir(ctx, session, dir, "git pull --ff-only")
}

// pullDir pulls one checkout under its lock, showing the command line.
func (m *Manager) pullDir(ctx context.Context, session *Session, dir, line string) error {
	lock := m.pullLock(dir)
	lock.Lock()
	defer lock.Unlock()

	fmt.Fprintf(session, "\x1b[2m$ %s\x1b[0m\r\n", line)
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

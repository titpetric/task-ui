package runner

import (
	"time"
)

// starting bounds how long a start waits for the session of another
// start of the same job, one that holds the lock while it writes its row.
const starting = 2 * time.Second

// startPoll is how often a start looks again for that session.
const startPoll = 5 * time.Millisecond

// acquire takes the lock of a key, killing the holder when forced, and
// returns the session in the way with ErrBusy when it may not. A nil
// holder is a start under way, whose session does not exist yet: the
// caller waits for it, briefly, and gets ErrBusy without a session past
// that. A folder that is moving takes none.
func (m *Manager) acquire(key, folder string, force bool) (*Session, error) {
	deadline := time.Now().Add(m.opts.KillTimeout)
	settled := time.Now().Add(starting)
	for {
		m.mu.Lock()
		if m.isFrozen(folder) {
			m.mu.Unlock()
			return nil, ErrMoving
		}
		current, busy := m.byKey[key]
		if !busy {
			m.byKey[key] = nil
			m.mu.Unlock()
			return nil, nil
		}
		m.mu.Unlock()
		if current == nil {
			if time.Now().After(settled) {
				return nil, ErrBusy
			}
			time.Sleep(startPoll)
			continue
		}
		if !force {
			return current, ErrBusy
		}
		current.Kill(m.opts.Grace)
		select {
		case <-current.Done():
		case <-time.After(time.Until(deadline)):
			return current, ErrBusy
		}
	}
}

// release frees the lock a start took and gave up before a session held
// it: the run that waits first for the key, if any, takes it and starts.
func (m *Manager) release(key string) {
	m.handOff(key, nil)
}

// forget drops a finished session from memory after a while.
func (m *Manager) forget(session *Session) {
	time.Sleep(m.opts.Keep)
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.sessions[session.ID()] == session {
		delete(m.sessions, session.ID())
	}
}

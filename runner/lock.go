package runner

import (
	"time"
)

// acquire takes the lock of a key, killing the holder when forced. A
// folder that is moving takes none.
func (m *Manager) acquire(key, folder string, force bool) (*Session, error) {
	deadline := time.Now().Add(m.opts.KillTimeout)
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

// release frees the lock of a key, when the session holds it.
func (m *Manager) release(key string, session *Session) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.byKey[key] == session {
		delete(m.byKey, key)
	}
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

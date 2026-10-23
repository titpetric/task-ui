package runner

import (
	"path/filepath"
	"strings"

	"github.com/titpetric/task-ui/model"
)

// Freeze holds the folder of an ID under the root still while its project moves: no run or pull until release.
// It refuses with ErrBusy while the folder, or one inside it, has an active run, and ErrMoving during another move.
func (m *Manager) Freeze(id string) (release func(), err error) {
	m.mu.Lock()
	if m.isFrozen(id) {
		m.mu.Unlock()
		return nil, ErrMoving
	}
	for key := range m.byKey {
		if strings.HasPrefix(key, id+"/") {
			m.mu.Unlock()
			return nil, ErrBusy
		}
	}
	if m.frozen == nil {
		m.frozen = map[string]bool{}
	}
	m.frozen[id] = true
	m.mu.Unlock()

	lock := m.pullLock(filepath.Join(m.opts.Root, filepath.FromSlash(id)))
	lock.Lock()
	return func() {
		lock.Unlock()
		m.mu.Lock()
		delete(m.frozen, id)
		m.mu.Unlock()
	}, nil
}

// isFrozen reports whether a move holds a folder, or one that holds it.
// The caller holds mu.
func (m *Manager) isFrozen(folder string) bool {
	for id := range m.frozen {
		if model.Under(folder, id) {
			return true
		}
	}
	return false
}

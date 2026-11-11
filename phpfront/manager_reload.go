package phpfront

import (
	"context"
)

// Reload drops every module and reads the roots again, so a module added to a
// checkout appears without a restart.
func (m *Manager) Reload(ctx context.Context) error {
	if m == nil {
		return nil
	}
	m.reset()
	return m.Start(ctx)
}

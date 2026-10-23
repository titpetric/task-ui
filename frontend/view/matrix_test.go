package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/config"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/permission"
)

// A matrix has a row per section, a cell per rule, and the application
// permissions after; a fixed one allows every cell.
func TestMatrix(t *testing.T) {
	t.Parallel()
	p := config.Default()
	policy := permission.Policy{"note/all": permission.Allow, "note/delete": permission.Deny}
	m := view.NewMatrix(p, policy, map[string]bool{"note/view": true}, false)
	assert.Len(t, m.Sections, len(p.Sections))
	assert.Len(t, m.Permissions, len(p.Extra)+len(p.Permissions))
	assert.Equal(t, len(p.Rules)+1, m.Columns())

	var notes view.MatrixRow
	for _, row := range m.Sections {
		if row.Key == "note" {
			notes = row
		}
	}
	assert.Equal(t, "Notes", notes.Title)
	assert.Equal(t, permission.Allow, notes.Cells[0].Effect)
	assert.False(t, notes.Cells[0].Checked, "no check names all")
	assert.Equal(t, "note/view", notes.Cells[1].Path)
	assert.True(t, notes.Cells[1].Allowed)
	assert.Equal(t, "", notes.Cells[1].Effect, "unset inherits")
	assert.Equal(t, permission.Deny, notes.Cells[3].Effect)
	assert.Equal(t, "key/read", m.Permissions[0].Path, "a section's own rule, a row of its own")
	assert.Equal(t, "Instance Keys, Allow Read Private", m.Permissions[0].Title)
	assert.Equal(t, "run", m.Permissions[1].Path)
	assert.False(t, m.Permissions[1].Allowed)

	fixed := view.NewMatrix(p, nil, nil, true)
	assert.True(t, fixed.Fixed)
	assert.Equal(t, permission.Allow, fixed.Permissions[0].Effect)
	assert.True(t, fixed.Sections[0].Cells[2].Allowed)
}

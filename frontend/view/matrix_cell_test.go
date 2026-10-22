package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/config"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/permission"
)

// A cell names its path with the section's and the rule's titles, and
// takes the effect the policy sets.
func TestMatrixCell(t *testing.T) {
	t.Parallel()
	m := view.NewMatrix(config.Default(), permission.Policy{"shell": permission.Deny}, nil, false)
	var shell view.MatrixCell
	for _, cell := range m.Permissions {
		if cell.Path == "shell" {
			shell = cell
		}
	}
	assert.Equal(t, "Enter Shell", shell.Title)
	assert.Equal(t, permission.Deny, shell.Effect)
	assert.True(t, shell.Checked)
	assert.Equal(t, "Workspaces, Allow View", m.Sections[0].Cells[1].Title)
}

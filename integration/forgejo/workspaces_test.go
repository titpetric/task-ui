package forgejo_test

import (
	"testing"

	"github.com/titpetric/task-ui/forgejo"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/storage"
)

// The workspace storage is the Workspaces sync writes.
func TestWorkspaces(t *testing.T) {
	t.Parallel()
	var rows forgejo.Workspaces = (*storage.WorkspaceStorage)(nil)
	assert.NotNil(t, &rows)
}

package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
)

// A home without a repository is empty, a nil home too.
func TestWorkspaceHome_Empty(t *testing.T) {
	t.Parallel()
	var none *view.WorkspaceHome
	assert.True(t, none.Empty())
	assert.True(t, (&view.WorkspaceHome{Folder: "malta/home"}).Empty())
	assert.False(t, (&view.WorkspaceHome{Folder: "malta/home", URL: "git@example.com:org/workspace.git"}).Empty())
}

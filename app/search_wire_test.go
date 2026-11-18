package app

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
)

// A folder ID names its workspace and project; a bare project is in internal.
func TestSplitFolderID(t *testing.T) {
	t.Parallel()
	workspace, project := splitFolderID("acme")
	assert.Equal(t, "internal", workspace)
	assert.Equal(t, "acme", project)

	workspace, project = splitFolderID("team/acme/docs")
	assert.Equal(t, "team", workspace)
	assert.Equal(t, "acme", project)
}

package search_test

import (
	"testing"

	"github.com/titpetric/task-ui/feature/search"
	"github.com/titpetric/task-ui/internal/assert"
)

// A source names a project, its directory.
func TestSource(t *testing.T) {
	t.Parallel()
	source := search.Source{Workspace: "internal", Project: "acme", FolderID: "acme", Dir: "/root/src/acme"}
	assert.Equal(t, "acme", source.FolderID)
	assert.Equal(t, "/root/src/acme", source.Dir)
}

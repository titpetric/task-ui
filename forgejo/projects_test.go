package forgejo_test

import (
	"testing"

	"github.com/titpetric/task-ui/forgejo"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/storage"
)

// The project storage is the Projects sync writes.
func TestProjects(t *testing.T) {
	t.Parallel()
	var rows forgejo.Projects = (*storage.ProjectStorage)(nil)
	assert.NotNil(t, &rows)
}

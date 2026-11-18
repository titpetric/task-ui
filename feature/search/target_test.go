package search_test

import (
	"testing"

	"github.com/titpetric/task-ui/feature/search"
	"github.com/titpetric/task-ui/internal/assert"
)

// A target names a job, its description.
func TestTarget(t *testing.T) {
	t.Parallel()
	target := search.Target{FolderID: "acme", ID: "ci:daily", Name: "ci:daily", Description: "Run the gate.", Steps: []string{"build"}}
	assert.Equal(t, "ci:daily", target.Name)
	assert.Equal(t, []string{"build"}, target.Steps)
}

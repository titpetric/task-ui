package forgejo_test

import (
	"testing"

	"github.com/titpetric/task-ui/integration/forgejo"
	"github.com/titpetric/task-ui/integration/git"
	"github.com/titpetric/task-ui/internal/assert"
)

// git.Git is the Git sync clones and pulls with.
func TestGit(t *testing.T) {
	t.Parallel()
	var g forgejo.Git = &git.Git{}
	assert.False(t, g.IsRepo(t.TempDir()))
}

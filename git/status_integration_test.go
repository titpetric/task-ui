//go:build integration

package git_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/git"
	"github.com/titpetric/task-ui/internal/assert"
)

func gitIn(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@t", "-c", "init.defaultBranch=main", "-c", "commit.gpgsign=false"}, args...)...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	assert.NoError(t, err, string(out))
}

// Status reads the nearest checkout up from a directory, no higher than
// the root: its branch, its commits ahead of and behind its upstream,
// and its changed files; a detached checkout names its commit.
func TestGit_Status(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	origin := filepath.Join(root, "origin")
	assert.NoError(t, os.MkdirAll(origin, 0o755))
	gitIn(t, origin, "init", "-q")
	assert.NoError(t, os.WriteFile(filepath.Join(origin, "a"), []byte("a\n"), 0o644))
	gitIn(t, origin, "add", "a")
	gitIn(t, origin, "commit", "-qm", "one")
	gitIn(t, root, "clone", "-q", "origin", "work")
	work := filepath.Join(root, "work")
	for _, n := range []string{"two", "three"} {
		assert.NoError(t, os.WriteFile(filepath.Join(origin, n), []byte(n), 0o644))
		gitIn(t, origin, "add", n)
		gitIn(t, origin, "commit", "-qm", n)
	}
	gitIn(t, work, "fetch", "-q")
	assert.NoError(t, os.WriteFile(filepath.Join(work, "b"), []byte("b\n"), 0o644))
	gitIn(t, work, "add", "b")
	gitIn(t, work, "commit", "-qm", "local")
	assert.NoError(t, os.WriteFile(filepath.Join(work, "a"), []byte("changed\n"), 0o644))
	assert.NoError(t, os.WriteFile(filepath.Join(work, "new"), []byte("new\n"), 0o644))
	inner := filepath.Join(work, "deep", "er")
	assert.NoError(t, os.MkdirAll(inner, 0o755))

	g := &git.Git{}
	s := g.Status(inner, root)
	assert.NotNil(t, s)
	assert.Equal(t, "work", s.Checkout)
	assert.Equal(t, "main", s.Branch)
	assert.Equal(t, "origin/main", s.Upstream)
	assert.Equal(t, 1, s.Ahead)
	assert.Equal(t, 2, s.Behind)
	assert.Equal(t, 2, s.Dirty)

	// The root bounds the search: from inside the checkout, a root
	// below it finds none.
	assert.Nil(t, g.Status(inner, filepath.Join(work, "deep")))
	assert.Nil(t, g.Status(root, root))
	assert.Nil(t, g.Status(filepath.Dir(root), root))

	detached := filepath.Join(root, "detached")
	gitIn(t, root, "clone", "-q", "origin", "detached")
	gitIn(t, detached, "checkout", "-q", "--detach", "HEAD~1")
	s = g.Status(detached, root)
	assert.NotNil(t, s)
	assert.Equal(t, "", s.Branch)
	assert.Len(t, s.Detached, 7)
	assert.Equal(t, 0, s.Dirty)
}

package skills_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/skills"
)

// Bin is the atkins the probe runs: Command, or atkins.
func TestProber_Bin(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "atkins", (&skills.Prober{}).Bin())
	assert.Equal(t, "atkins", (*skills.Prober)(nil).Bin())
	assert.Equal(t, "/opt/atkins", (&skills.Prober{Command: "/opt/atkins"}).Bin())
}

// Stamp changes with the nearest pipeline file and the files of every
// skills folder from the directory up.
func TestStamp(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	inner := filepath.Join(root, "a", "b")
	assert.NoError(t, os.MkdirAll(inner, 0o755))
	before := skills.Stamp(inner)
	assert.NoError(t, os.MkdirAll(filepath.Join(root, ".atkins", "skills"), 0o755))
	withSkills := skills.Stamp(inner)
	assert.NotEqual(t, before, withSkills)
	assert.NoError(t, os.WriteFile(filepath.Join(root, "a", "atkins.yml"), []byte("jobs: {}\n"), 0o644))
	assert.NotEqual(t, withSkills, skills.Stamp(inner))
	assert.Contains(t, skills.Stamp(inner), "atkins.yml@"+filepath.Join(root, "a"))
}

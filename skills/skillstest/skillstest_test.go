package skillstest_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/skills/skillstest"
)

// The stand-in lists git from the nearest checkout and compose beside a
// compose file, compose:build only when the file builds.
func TestProber(t *testing.T) {
	t.Parallel()
	p, calls := skillstest.Prober(t)
	root, err := filepath.EvalSymlinks(t.TempDir())
	assert.NoError(t, err)
	app := filepath.Join(root, "app")
	assert.NoError(t, os.MkdirAll(filepath.Join(root, ".git"), 0o755))
	assert.NoError(t, os.MkdirAll(app, 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(app, "compose.yml"), []byte("services:\n  a:\n    image: x\n"), 0o644))

	sections := p.Probe(app)
	assert.Len(t, sections, 3)
	assert.Equal(t, root, sections[0].Dir)
	assert.Equal(t, app, sections[1].Dir)
	assert.False(t, sections[1].Cmds[0].Active)
	assert.True(t, sections[1].Cmds[1].Args)
	assert.False(t, sections[2].Active)
	data, err := os.ReadFile(calls)
	assert.NoError(t, err)
	assert.Equal(t, app+"\n", string(data))
}

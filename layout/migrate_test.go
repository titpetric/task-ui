package layout_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/layout"
)

// Plan moves a workspace's home project to home and its other
// directories under src, the same at the root with the workspaces left
// in place, makes an empty home for a workspace without a home project, and once
// applied plans nothing more.
func TestPlan(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for _, dir := range []string{
		"apps", "deploy", "mobius", "workspace", ".git",
		"malta/workspace", "malta/apps", "malta/dockerfiles", "malta/.cache",
		"internal/tools",
	} {
		assert.NoError(t, os.MkdirAll(filepath.Join(root, filepath.FromSlash(dir)), 0o755))
	}
	assert.NoError(t, os.WriteFile(filepath.Join(root, "notes.txt"), []byte("a file at the root stays\n"), 0o644))
	assert.NoError(t, os.WriteFile(filepath.Join(root, "malta", "workspace", "CLAUDE.md"), []byte("# malta\n"), 0o644))

	moves, err := layout.Plan(root, []string{"malta"}, map[string]string{"malta": "workspace"})
	assert.NoError(t, err)
	lines := []string{}
	for _, move := range moves {
		lines = append(lines, move.String())
	}
	assert.Equal(t, []string{
		"mkdir " + filepath.Join(root, "home"),
		"mkdir " + filepath.Join(root, "malta", "src"),
		"mkdir " + filepath.Join(root, "src"),
		"mv " + filepath.Join(root, "malta", "workspace") + " " + filepath.Join(root, "malta", "home"),
		"mv " + filepath.Join(root, "malta", "apps") + " " + filepath.Join(root, "malta", "src", "apps"),
		"mv " + filepath.Join(root, "malta", "dockerfiles") + " " + filepath.Join(root, "malta", "src", "dockerfiles"),
		"mv " + filepath.Join(root, "apps") + " " + filepath.Join(root, "src", "apps"),
		"mv " + filepath.Join(root, "deploy") + " " + filepath.Join(root, "src", "deploy"),
		"mv " + filepath.Join(root, "mobius") + " " + filepath.Join(root, "src", "mobius"),
		"mv " + filepath.Join(root, "workspace") + " " + filepath.Join(root, "src", "workspace"),
	}, lines, "the root has no home project: workspace is an ordinary checkout")

	assert.NoError(t, layout.Apply(moves))
	assert.FileExists(t, filepath.Join(root, "malta", "home", "CLAUDE.md"))
	assert.DirExists(t, filepath.Join(root, "malta", "src", "dockerfiles"))
	assert.DirExists(t, filepath.Join(root, "malta", ".cache"), "dot directories stay")
	assert.DirExists(t, filepath.Join(root, "src", "workspace"))
	assert.DirExists(t, filepath.Join(root, "internal", "tools"), "task-ui's own directory stays")
	assert.DirExists(t, filepath.Join(root, "home"))
	entries, err := os.ReadDir(filepath.Join(root, "home"))
	assert.NoError(t, err)
	assert.Len(t, entries, 0, "the root's home is empty")
	assert.NoDirExists(t, filepath.Join(root, "malta", "workspace"))
	assert.FileExists(t, filepath.Join(root, "notes.txt"))

	again, err := layout.Plan(root, []string{"malta"}, map[string]string{"malta": "workspace"})
	assert.NoError(t, err)
	assert.Len(t, again, 0, "laid out already")
}

// A directory that exists on both sides of a move is an error.
func TestPlan_Conflict(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for _, dir := range []string{"apps", "src/apps", "home"} {
		assert.NoError(t, os.MkdirAll(filepath.Join(root, filepath.FromSlash(dir)), 0o755))
	}
	_, err := layout.Plan(root, nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "src/apps")
}

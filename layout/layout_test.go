package layout_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/layout"
)

// A folder ID maps to its directory and back: a project of _ under
// <root>/src, a project of a named workspace under <root>/<ws>/src, the
// home of each at <root>[/<ws>]/home, and a folder inside a project
// under the project.
func TestDirAndID(t *testing.T) {
	t.Parallel()
	named := []string{"malta"}
	root := filepath.FromSlash("/data")
	for _, c := range []struct{ id, dir string }{
		{"_", "/data/src"},
		{"apps", "/data/src/apps"},
		{"apps/blog", "/data/src/apps/blog"},
		{"home", "/data/home"},
		{"home/.atkins", "/data/home/.atkins"},
		{"malta", "/data/malta/src"},
		{"malta/dockerfiles", "/data/malta/src/dockerfiles"},
		{"malta/dockerfiles/task", "/data/malta/src/dockerfiles/task"},
		{"malta/home", "/data/malta/home"},
		{"malta/home/skills", "/data/malta/home/skills"},
	} {
		assert.Equal(t, filepath.FromSlash(c.dir), layout.Dir(root, named, c.id), c.id)
		id, ok := layout.ID(root, named, filepath.FromSlash(c.dir))
		assert.True(t, ok, c.dir)
		assert.Equal(t, c.id, id, c.dir)
	}
	for _, dir := range []string{"/data", "/data/malta", "/elsewhere", "/data/notes", "/data/malta/notes"} {
		_, ok := layout.ID(root, named, filepath.FromSlash(dir))
		assert.False(t, ok, dir)
	}
}

func TestDirs(t *testing.T) {
	t.Parallel()
	root := filepath.FromSlash("/data")
	assert.Equal(t, root, layout.WorkspaceDir(root, "_"))
	assert.Equal(t, filepath.FromSlash("/data/malta"), layout.WorkspaceDir(root, "malta"))
	assert.Equal(t, filepath.FromSlash("/data/home"), layout.HomeDir(root, "_"))
	assert.Equal(t, filepath.FromSlash("/data/malta/home"), layout.HomeDir(root, "malta"))
	assert.Equal(t, filepath.FromSlash("/data/src"), layout.SrcDir(root, ""))
	assert.Equal(t, filepath.FromSlash("/data/malta/src/apps/blog"), layout.ProjectDir(root, "malta", "apps/blog"))
	assert.Equal(t, "home", layout.HomeID("_"))
	assert.Equal(t, "malta/home", layout.HomeID("malta"))
}

// Split and Join tell a workspace apart from a project of _ by the
// named workspaces, and IsHome names the home folders and what is in them.
func TestSplitJoin(t *testing.T) {
	t.Parallel()
	named := []string{"malta"}
	ws, rel := layout.Split(named, "malta/dockerfiles/task")
	assert.Equal(t, "malta", ws)
	assert.Equal(t, "dockerfiles/task", rel)
	ws, rel = layout.Split(named, "apps/blog")
	assert.Equal(t, "_", ws)
	assert.Equal(t, "apps/blog", rel)
	ws, rel = layout.Split(named, "malta")
	assert.Equal(t, "malta", ws)
	assert.Equal(t, "", rel)
	ws, rel = layout.Split(named, "_")
	assert.Equal(t, "_", ws)
	assert.Equal(t, "", rel)
	assert.Equal(t, "_", layout.Join("_", ""))
	assert.Equal(t, "apps", layout.Join("_", "apps"))
	assert.Equal(t, "malta", layout.Join("malta", ""))
	assert.Equal(t, "malta/apps", layout.Join("malta", "apps"))
	assert.True(t, layout.IsHome(named, "home"))
	assert.True(t, layout.IsHome(named, "malta/home/skills"))
	assert.False(t, layout.IsHome(named, "malta/homes"))
	assert.False(t, layout.IsHome(named, "apps/home"))
	assert.True(t, layout.Reserved("home"))
	assert.True(t, layout.Reserved("src"))
	assert.True(t, layout.Reserved("_"))
	assert.False(t, layout.Reserved("homes"))
}

// Workspace names the workspace a directory is in, and Ensure makes the
// two directories of one.
func TestWorkspaceEnsure(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	assert.NoError(t, layout.Ensure(root, "_"))
	assert.NoError(t, layout.Ensure(root, "malta"))
	for _, dir := range []string{"home", "src", "malta/home", "malta/src"} {
		info, err := os.Stat(filepath.Join(root, filepath.FromSlash(dir)))
		assert.NoError(t, err, dir)
		assert.True(t, info.IsDir(), dir)
	}
	ws, ok := layout.Workspace(root, []string{"malta"}, filepath.Join(root, "malta", "src", "apps"))
	assert.True(t, ok)
	assert.Equal(t, "malta", ws)
	ws, ok = layout.Workspace(root, []string{"malta"}, filepath.Join(root, "src", "apps"))
	assert.True(t, ok)
	assert.Equal(t, "_", ws)
	_, ok = layout.Workspace(root, nil, filepath.Dir(root))
	assert.False(t, ok)
}

// ATKINS_DB names the home's history first and the instance's mirror second.
func TestAtkinsDB(t *testing.T) {
	t.Parallel()
	sep := string(os.PathListSeparator)
	assert.Equal(t, filepath.FromSlash("/data/malta/home/.atkins/atkins.db")+sep+filepath.FromSlash("/root/.atkins/atkins.db"), layout.AtkinsDB(filepath.FromSlash("/data/malta/home"), filepath.FromSlash("/root")))
	assert.Equal(t, filepath.FromSlash("/data/home/.atkins/atkins.db"), layout.AtkinsDB(filepath.FromSlash("/data/home"), ""))
}

package folder_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/internal/folder"
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
		{"internal", "/data/src"},
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
		assert.Equal(t, filepath.FromSlash(c.dir), folder.Dir(root, named, c.id), c.id)
		id, ok := folder.ID(root, named, filepath.FromSlash(c.dir))
		assert.True(t, ok, c.dir)
		assert.Equal(t, c.id, id, c.dir)
	}
	for _, dir := range []string{"/data", "/data/malta", "/elsewhere", "/data/notes", "/data/malta/notes"} {
		_, ok := folder.ID(root, named, filepath.FromSlash(dir))
		assert.False(t, ok, dir)
	}
}

func TestDirs(t *testing.T) {
	t.Parallel()
	root := filepath.FromSlash("/data")
	assert.Equal(t, root, folder.WorkspaceDir(root, "internal"))
	assert.Equal(t, filepath.FromSlash("/data/malta"), folder.WorkspaceDir(root, "malta"))
	assert.Equal(t, filepath.FromSlash("/data/home"), folder.HomeDir(root, "internal"))
	assert.Equal(t, filepath.FromSlash("/data/malta/home"), folder.HomeDir(root, "malta"))
	assert.Equal(t, filepath.FromSlash("/data/src"), folder.SrcDir(root, ""))
	assert.Equal(t, filepath.FromSlash("/data/malta/src/apps/blog"), folder.ProjectDir(root, "malta", "apps/blog"))
	assert.Equal(t, "home", folder.HomeID("internal"))
	assert.Equal(t, "malta/home", folder.HomeID("malta"))
}

// Split and Join tell a workspace apart from a project of _ by the
// named workspaces, and IsHome names the home folders and what is in them.
func TestSplitJoin(t *testing.T) {
	t.Parallel()
	named := []string{"malta"}
	ws, rel := folder.Split(named, "malta/dockerfiles/task")
	assert.Equal(t, "malta", ws)
	assert.Equal(t, "dockerfiles/task", rel)
	ws, rel = folder.Split(named, "apps/blog")
	assert.Equal(t, "internal", ws)
	assert.Equal(t, "apps/blog", rel)
	ws, rel = folder.Split(named, "malta")
	assert.Equal(t, "malta", ws)
	assert.Equal(t, "", rel)
	ws, rel = folder.Split(named, "internal")
	assert.Equal(t, "internal", ws)
	assert.Equal(t, "", rel)
	assert.Equal(t, "internal", folder.Join("internal", ""))
	assert.Equal(t, "apps", folder.Join("internal", "apps"))
	assert.Equal(t, "malta", folder.Join("malta", ""))
	assert.Equal(t, "malta/apps", folder.Join("malta", "apps"))
	assert.True(t, folder.IsHome(named, "home"))
	assert.True(t, folder.IsHome(named, "malta/home/skills"))
	assert.False(t, folder.IsHome(named, "malta/homes"))
	assert.False(t, folder.IsHome(named, "apps/home"))
	assert.True(t, folder.Reserved("home"))
	assert.True(t, folder.Reserved("src"))
	assert.True(t, folder.Reserved("internal"))
	assert.True(t, folder.Reserved("_"))
	assert.False(t, folder.Reserved("homes"))
}

// Workspace names the workspace a directory is in, and Ensure makes the
// two directories of one.
func TestWorkspaceEnsure(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	assert.NoError(t, folder.Ensure(root, "internal"))
	assert.NoError(t, folder.Ensure(root, "malta"))
	for _, dir := range []string{"home", "src", "malta/home", "malta/src"} {
		info, err := os.Stat(filepath.Join(root, filepath.FromSlash(dir)))
		assert.NoError(t, err, dir)
		assert.True(t, info.IsDir(), dir)
	}
	ws, ok := folder.Workspace(root, []string{"malta"}, filepath.Join(root, "malta", "src", "apps"))
	assert.True(t, ok)
	assert.Equal(t, "malta", ws)
	ws, ok = folder.Workspace(root, []string{"malta"}, filepath.Join(root, "src", "apps"))
	assert.True(t, ok)
	assert.Equal(t, "internal", ws)
	_, ok = folder.Workspace(root, nil, filepath.Dir(root))
	assert.False(t, ok)
}

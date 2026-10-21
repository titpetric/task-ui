package folder_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
)

// checkout is a folder with a file, an executable, a sealed directory
// and file, and symlinks inside and out of it.
func checkout(t *testing.T) *folder.Folder {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "api")
	outside := filepath.Join(filepath.Dir(dir), "secret.txt")
	assert.NoError(t, os.MkdirAll(filepath.Join(dir, ".git"), 0o755))
	assert.NoError(t, os.MkdirAll(filepath.Join(dir, "bin"), 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, ".git", "config"), []byte("[core]\n"), 0o644))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, ".env"), []byte("A=1\n"), 0o644))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("# api\n"), 0o644))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "bin", "deploy"), []byte("#!/bin/sh\necho deployed\n"), 0o755))
	assert.NoError(t, os.WriteFile(outside, []byte("nope\n"), 0o644))
	assert.NoError(t, os.Symlink(outside, filepath.Join(dir, "escape")))
	assert.NoError(t, os.Symlink("README.md", filepath.Join(dir, "readme")))
	assert.NoError(t, os.Symlink(".git/config", filepath.Join(dir, "config")))
	return &folder.Folder{ID: "acme/api", Dir: dir}
}

// Path opens what is inside the folder and nothing else: no .., no
// symlink out, nothing sealed or through a seal, even by a symlink.
func TestFolder_Path(t *testing.T) {
	t.Parallel()
	f := checkout(t)
	for _, rel := range []string{"", "README.md", "bin", "bin/deploy", "readme", "./README.md"} {
		_, ok := f.Path(rel)
		assert.True(t, ok, rel)
	}
	for _, rel := range []string{"../secret.txt", "bin/../README.md", "escape", ".git", ".git/config", ".env", "config", "/etc/passwd", "missing"} {
		_, ok := f.Path(rel)
		assert.False(t, ok, rel)
	}
}

// A file that runs is a job of its folder by its path, ./ in front.
func TestFolder_FileJob(t *testing.T) {
	t.Parallel()
	f := checkout(t)
	job := f.Job("./bin/deploy")
	assert.NotNil(t, job)
	assert.Equal(t, "./bin/deploy", job.ID)
	assert.Equal(t, "acme/api/./bin/deploy", job.Key())
	assert.Equal(t, "file", job.Kind)
	assert.Equal(t, []string{"./bin/deploy"}, job.Command)
	assert.Nil(t, f.Job("./README.md"))
	assert.Nil(t, f.Job("./bin"))
	assert.Nil(t, f.Job("./bin/../bin/deploy"))
	assert.Nil(t, f.Job("./escape"))
}

// A listing marks the files that run and lists a symlink as what it
// points to.
func TestListing_Exec(t *testing.T) {
	t.Parallel()
	f := checkout(t)
	exec := map[string]bool{}
	for _, item := range folder.Listing(f.Dir) {
		exec[item.Name] = item.Exec
	}
	assert.False(t, exec["README.md"])
	assert.False(t, exec["readme"])
	assert.False(t, exec["bin"])
	items := folder.Listing(filepath.Join(f.Dir, "bin"))
	assert.Len(t, items, 1)
	assert.True(t, items[0].Exec)
}

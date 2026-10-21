package folder_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
)

// checkout is a folder with a file, an executable, .git, dot files and
// dot directories, and symlinks inside and out of it.
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
	for name, body := range map[string]string{".claude/CLAUDE.md": "# claude\n", ".claude/.secret": "s\n", ".atkins/skills/x.yml": "x: 1\n", ".gitignore": "x\n"} {
		assert.NoError(t, os.MkdirAll(filepath.Join(dir, filepath.Dir(name)), 0o755))
		assert.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644))
	}
	assert.NoError(t, os.WriteFile(filepath.Join(dir, ".hook"), []byte("#!/bin/sh\necho hook\n"), 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, ".claude", ".run"), []byte("#!/bin/sh\necho run\n"), 0o755))
	assert.NoError(t, os.Symlink(".env", filepath.Join(dir, "envlink")))
	assert.NoError(t, os.Symlink(".hook", filepath.Join(dir, "hook")))
	assert.NoError(t, os.Symlink("README.md", filepath.Join(dir, ".readme")))
	assert.NoError(t, os.Symlink(".claude", filepath.Join(dir, "claude")))
	assert.NoError(t, os.Symlink(".git", filepath.Join(dir, "repo")))
	assert.NoError(t, os.Symlink("nowhere", filepath.Join(dir, ".gone")))
	return &folder.Folder{ID: "acme/api", Dir: dir}
}

// Path opens what is inside the folder and nothing else: no .., no
// symlink out, no dot file, nothing in .git, nor any of them by a
// symlink. A dot directory other than .git opens, and what is in it
// but its dot files.
func TestFolder_Path(t *testing.T) {
	t.Parallel()
	f := checkout(t)
	for _, rel := range []string{"", "README.md", "bin", "bin/deploy", "readme", "./README.md",
		".claude", ".claude/CLAUDE.md", ".atkins/skills/x.yml", "claude", "claude/CLAUDE.md"} {
		_, ok := f.Path(rel)
		assert.True(t, ok, rel)
	}
	for _, rel := range []string{"../secret.txt", "bin/../README.md", "escape", ".git", ".git/config", ".env", "config", "/etc/passwd", "missing",
		".gitignore", ".hook", ".claude/.secret", ".claude/.run", "claude/.secret", "envlink", "hook", ".readme", "repo", "repo/config", ".gone"} {
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
	for _, id := range []string{"./.hook", "./hook", "./.claude/.run", "./claude/.run"} {
		assert.Nil(t, f.Job(id), id)
	}
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

// A listing seals .git and the dot files, a symlink to either and a
// dangling dot link, and no dot directory; a sealed file does not run.
func TestListing_Sealed(t *testing.T) {
	t.Parallel()
	f := checkout(t)
	sealed, exec := map[string]bool{}, map[string]bool{}
	for _, item := range folder.Listing(f.Dir) {
		sealed[item.Name], exec[item.Name] = item.Sealed, item.Exec
	}
	for _, name := range []string{".git", ".env", ".gitignore", ".hook", ".readme", ".gone", "envlink", "hook", "config", "repo"} {
		assert.True(t, sealed[name], name)
	}
	for _, name := range []string{".claude", ".atkins", "claude", "bin", "README.md", "readme", "escape"} {
		assert.False(t, sealed[name], name)
	}
	assert.False(t, exec[".hook"])
	assert.False(t, exec["hook"])
	inner := map[string]bool{}
	for _, item := range folder.Listing(filepath.Join(f.Dir, ".claude")) {
		inner[item.Name] = item.Sealed
		assert.False(t, item.Exec, item.Name)
	}
	assert.Equal(t, map[string]bool{"CLAUDE.md": false, ".secret": true, ".run": true}, inner)
}

// Sub reads a directory inside the folder as Load does, its jobs keyed
// and run by that directory under the folder's ID and rules; a path
// Path refuses, a file, or a path out by a symlink is none.
func TestFolder_Sub(t *testing.T) {
	t.Parallel()
	f := checkout(t)
	f.Allow, f.Hours = []string{"ops"}, "Mon-Fri 08-18"
	images := filepath.Join(f.Dir, "images")
	assert.NoError(t, os.MkdirAll(images, 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(images, "atkins.yml"), []byte("jobs:\n  default:\n    desc: All\n    steps:\n      - run: echo all\n  custom:\n    desc: Custom\n    args: true\n    steps:\n      - run: echo\n"), 0o644))
	assert.NoError(t, os.Symlink(filepath.Dir(f.Dir), filepath.Join(f.Dir, "out")))

	assert.Equal(t, f, f.Sub(""))
	sub := f.Sub("images/")
	assert.NotNil(t, sub)
	assert.Equal(t, "acme/api", sub.ID)
	assert.Equal(t, "images", sub.Rel)
	assert.Equal(t, []string{"ops"}, sub.Allow)
	assert.Equal(t, "Mon-Fri 08-18", sub.Hours)
	job := sub.Job("default")
	assert.NotNil(t, job)
	assert.Equal(t, "images", job.Dir)
	assert.Equal(t, "acme/api/images/default", job.Key())
	assert.Equal(t, []string{"atkins", "-f", "atkins.yml", "default"}, job.Command)
	assert.True(t, sub.Job("custom").TakesArgs)
	assert.False(t, job.TakesArgs)

	for _, rel := range []string{"..", "images/..", "./images", ".git", "repo", "escape", "out", "README.md", "missing", "/etc"} {
		assert.Nil(t, f.Sub(rel), rel)
	}
	assert.Nil(t, (*folder.Folder)(nil).Sub("images"))
}

package phpfront_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/phpfront"
)

// write makes a file and every directory above it.
func write(t *testing.T, path, body string) string {
	t.Helper()
	assert.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	assert.NoError(t, os.WriteFile(path, []byte(body), 0o644))
	return path
}

const startup = "<?php\n// @startup\nTask::register(new Task\\Module);\n"

func TestFound(t *testing.T) {
	t.Parallel()
	entry := phpfront.Found{Root: "/data/modules/titpetric/refinement", File: "/data/modules/titpetric/refinement/modinfo.php", Startup: true}
	assert.Equal(t, "/data/modules/titpetric/refinement", entry.Root)
	assert.Equal(t, filepath.Join(entry.Root, phpfront.ModInfo), entry.File)
	assert.True(t, entry.Startup)
}

func TestDiscover(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	write(t, filepath.Join(root, "titpetric", "refinement", phpfront.ModInfo), startup)
	write(t, filepath.Join(root, "acme", "notes", phpfront.ModInfo), "<?php\n// no annotation\n")
	write(t, filepath.Join(root, "acme", "notes", "vendor", "dep", phpfront.ModInfo), startup)
	write(t, filepath.Join(root, "acme", "notes", "node_modules", "dep", phpfront.ModInfo), startup)
	write(t, filepath.Join(root, "acme", "notes", "index.php"), "<?php\n")

	found, err := phpfront.Discover(root)
	assert.NoError(t, err)
	assert.Len(t, found, 2, "a vendor or node_modules tree carries no module")

	assert.Equal(t, filepath.Join(root, "acme", "notes"), found[0].Root, "the roots come back by path")
	assert.False(t, found[0].Startup, "a modinfo.php with no @startup is reported, not run")
	assert.Equal(t, filepath.Join(root, "titpetric", "refinement"), found[1].Root)
	assert.True(t, found[1].Startup)
	assert.Equal(t, filepath.Join(found[1].Root, phpfront.ModInfo), found[1].File)
}

func TestDiscoverRoots(t *testing.T) {
	t.Parallel()
	first, second := t.TempDir(), t.TempDir()
	write(t, filepath.Join(first, "a", "one", phpfront.ModInfo), startup)
	write(t, filepath.Join(second, "b", "two", phpfront.ModInfo), startup)

	found, err := phpfront.Discover(first, second, "", filepath.Join(first, "gone"))
	assert.NoError(t, err)
	assert.Len(t, found, 2, "a root that is not there is an empty root")

	// The same root twice registers the module once.
	found, err = phpfront.Discover(first, first)
	assert.NoError(t, err)
	assert.Len(t, found, 1)

	found, err = phpfront.Discover()
	assert.NoError(t, err)
	assert.Len(t, found, 0)
}

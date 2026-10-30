package phpscripts_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/phpscripts"
)

func TestModules(t *testing.T) {
	t.Parallel()
	tree := phpscripts.Modules()

	// The README of the tree is at its root, so the first segment of a path
	// in it is a vendor and nothing above it.
	_, err := fs.Stat(tree, "README.md")
	assert.NoError(t, err)

	_, err = fs.Stat(tree, "modules")
	assert.Error(t, err, "the tree is rooted at modules, not above it")

	_, err = fs.Stat(tree, filepath.Join("titpetric", "refinement", "modinfo.php"))
	assert.NoError(t, err, "the refinement module ships in the binary")
}

func TestExtract(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "modules")
	assert.NoError(t, phpscripts.Extract(dir))

	shipped, err := fs.ReadFile(phpscripts.Modules(), "README.md")
	assert.NoError(t, err)
	written, err := os.ReadFile(filepath.Join(dir, "README.md"))
	assert.NoError(t, err)
	assert.Equal(t, string(shipped), string(written))

	info, err := os.Stat(filepath.Join(dir, "README.md"))
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0o644), info.Mode().Perm(), "nothing in the tree is executable")

	// composer's tree travels with the module: a running instance installs
	// nothing and reaches no network.
	assert.FileExists(t, filepath.Join(dir, "titpetric", "refinement", "vendor", "autoload.php"))
}

func TestExtractReplaces(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("stale"), 0o644))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "kept.txt"), []byte("mine"), 0o644))

	assert.NoError(t, phpscripts.Extract(dir))

	written, err := os.ReadFile(filepath.Join(dir, "README.md"))
	assert.NoError(t, err)
	assert.NotEqual(t, "stale", string(written), "an upgrade replaces a shipped file")

	kept, err := os.ReadFile(filepath.Join(dir, "kept.txt"))
	assert.NoError(t, err)
	assert.Equal(t, "mine", string(kept), "what the instance wrote beside the shipped files stays")
}

func TestExtractRefuses(t *testing.T) {
	t.Parallel()
	file := filepath.Join(t.TempDir(), "notadir")
	assert.NoError(t, os.WriteFile(file, []byte("x"), 0o644))
	assert.Error(t, phpscripts.Extract(file), "a target that is a file is not a modules directory")
}

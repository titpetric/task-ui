package phpscripts_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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

	// MiniTPL travels with the module: a running instance installs nothing
	// and reaches no network.
	assert.FileExists(t, filepath.Join(dir, "titpetric", "refinement", "lib", "MiniTPL", "Template.php"))
}

// No file a module needs may sit under a directory named vendor.
//
// The tree is embedded from a checkout, where a vendor directory is a
// directory like any other, and it is read from a release, where it is not:
// the binary is installed with `go install <module>@<ref>`, the Go module
// proxy builds that zip, and the zip drops what sits under vendor. A module
// that reaches into one works in every test here and fails on the instance
// with "no such file or directory", which is how it shipped once.
func TestModulesCarryNoVendor(t *testing.T) {
	t.Parallel()
	assert.NoError(t, fs.WalkDir(phpscripts.Modules(), ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		for _, part := range strings.Split(name, "/") {
			assert.NotEqual(t, "vendor", part, name+" sits under a vendor directory, which the module zip of a release drops")
		}
		return nil
	}))
}

// requirePath is the path of a require in a module's PHP, which the runtime
// resolves against the module's root.
var requirePath = regexp.MustCompile(`require(?:_once)?\s+"([^"]+)"`)

// Every require of every shipped PHP file names a file that shipped with it.
//
// The runtime roots a module at its own directory, so a require is a path
// inside the module and nothing above it resolves. A require that names a
// file the tree does not hold is a 500 on the first request, and the only
// place to catch it before the instance is here.
func TestModulesRequiresResolve(t *testing.T) {
	t.Parallel()
	tree := phpscripts.Modules()
	checked := 0
	assert.NoError(t, fs.WalkDir(tree, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || filepath.Ext(name) != ".php" {
			return err
		}
		// The first two segments are the module's vendor and plugin; a
		// require is relative to the directory they name.
		parts := strings.SplitN(name, "/", 3)
		if len(parts) < 3 {
			return nil
		}
		root := parts[0] + "/" + parts[1]
		source, err := fs.ReadFile(tree, name)
		if err != nil {
			return err
		}
		for _, match := range requirePath.FindAllStringSubmatch(string(source), -1) {
			checked++
			_, err := fs.Stat(tree, root+"/"+match[1])
			assert.NoError(t, err, name+" requires "+match[1]+", which did not ship")
		}
		return nil
	}))
	assert.True(t, checked > 0, "the shipped modules require something")
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

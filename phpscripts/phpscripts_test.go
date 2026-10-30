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

	// composer's tree travels with the module: a running instance installs
	// nothing and reaches no network.
	assert.FileExists(t, filepath.Join(dir, "titpetric", "refinement", "ext", "autoload.php"))
	assert.FileExists(t, filepath.Join(dir, "titpetric", "refinement", "ext", "composer", "autoload_real.php"))
	assert.FileExists(t, filepath.Join(dir, "titpetric", "refinement", "ext", "titpetric", "minitpl", "code", "MiniTPL", "Template.php"))
}

// A module installs its PHP dependencies into ext, never vendor.
//
// The tree is embedded from a checkout, where a vendor directory is a
// directory like any other, and read from a release, where it is not. The
// binary is installed with `go install <module>@<ref>`, the Go module proxy
// builds that zip, and x/mod/zip drops every file more than one level under
// a directory named vendor: isVendoredPackage finds the "vendor/" element
// and asks whether a slash follows it.
//
// A composer tree there arrives as one file, vendor/autoload.php, with
// nothing left for it to load. Every test here still passes, because they
// read the checkout, and the instance answers the first request with "open
// vendor/composer/autoload_real.php: no such file or directory", which is
// how it shipped once. ext is an ordinary directory to Go and to composer
// both: composer.json names it in config.vendor-dir.
func TestModulesInstallIntoExt(t *testing.T) {
	t.Parallel()
	assert.NoError(t, fs.WalkDir(phpscripts.Modules(), ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		for _, part := range strings.Split(name, "/") {
			assert.NotEqual(t, "vendor", part, name+": a module installs into ext, because the module zip of a release drops what is under vendor")
		}
		return nil
	}))
}

// requirePath is the path of a require in a module's PHP, which the runtime
// resolves against the module's root.
var requirePath = regexp.MustCompile(`require(?:_once)?\s+"([^"]+)"`)

// Every require a module writes names a file that shipped with it.
//
// A module's own files are entered from its root, so the path in a require
// is read from there: index.php is the root's own file, stubs/autoload.php
// arrives through `phpscript lint --include` from the root, and the fixtures
// of tests/ run under the suite root phpscript.yml puts there. A require
// that names a file the tree does not hold is a 500 on the first request,
// and the only place to catch it before the instance is here.
//
// ext/ is left out: it is what composer installed, its files find each other
// through __DIR__ and the autoloader, and upstream's paths are not ours to
// resolve.
func TestModulesRequiresResolve(t *testing.T) {
	t.Parallel()
	tree := phpscripts.Modules()
	checked := 0
	assert.NoError(t, fs.WalkDir(tree, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || filepath.Ext(name) != ".php" {
			return err
		}
		// The first two segments are the module's vendor and plugin.
		parts := strings.SplitN(name, "/", 3)
		if len(parts) < 3 || strings.HasPrefix(parts[2], "ext/") {
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

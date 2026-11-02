package phpfront_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/phpfront"
)

// The embedded tree holds the application: the autoloader, the domain, the
// theme and the tools. A release that shipped without one of them would answer
// its first request with a missing file.
func TestApp(t *testing.T) {
	t.Parallel()

	for _, name := range []string{
		"composer.json", "composer.lock", "ext/autoload.php", "phpscript.yml",
		"code/Module.php", "code/Task/Row.php", "code/Task/Model/Workspace.php",
		"stubs/autoload.php", "stubs/host.php",
		"tools/models.php", "tools/templates/model.tpl",
		"theme/assets/task-ui.css", "theme/assets/task-ui.js",
	} {
		data, err := fs.ReadFile(phpfront.App(), name)
		assert.NoError(t, err, name)
		assert.NotEmpty(t, data, name)
	}
}

// composer installs into ext, because x/mod/zip drops a file more than one
// level under a directory named vendor and the autoload tree is three deep: a
// release would ship vendor/autoload.php with nothing left for it to load.
func TestApp_InstallsIntoExt(t *testing.T) {
	t.Parallel()

	_, err := fs.ReadFile(phpfront.App(), "ext/composer/autoload_real.php")
	assert.NoError(t, err)
	_, err = fs.Stat(phpfront.App(), "vendor")
	assert.Error(t, err)
}

// The same rule holds for the browser libraries under theme/assets/vendor:
// one level survives the zip and a subdirectory does not.
func TestApp_VendorIsOneLevelDeep(t *testing.T) {
	t.Parallel()

	assert.NoError(t, fs.WalkDir(phpfront.App(), ".", func(name string, entry fs.DirEntry, err error) error {
		assert.NoError(t, err, name)
		if entry.IsDir() || !strings.Contains(name, "vendor/") {
			return nil
		}
		_, under, _ := strings.Cut(name, "vendor/")
		assert.NotContains(t, under, "/", name+": x/mod/zip drops this from a release")
		return nil
	}))
}

// The theme's assets come out of the embed, so /static/ serves one copy of the
// stylesheet while the Go pages are still there.
func TestThemeAssets(t *testing.T) {
	t.Parallel()

	data, err := fs.ReadFile(phpfront.ThemeAssets(), "task-ui.css")
	assert.NoError(t, err)
	assert.Contains(t, string(data), "@import url(")
}

// Extract writes the tree where a runtime can read and write it. Setenv and
// t.Parallel do not mix, so these three run in order.
func TestExtract(t *testing.T) {
	t.Setenv(phpfront.Env, "")

	root, err := phpfront.Extract(t.TempDir())
	assert.NoError(t, err)
	assert.Equal(t, phpfront.Dir, filepath.Base(root))
	for _, name := range []string{"composer.json", "ext/autoload.php", "code/Task/Row.php"} {
		_, err := os.Stat(filepath.Join(root, name))
		assert.NoError(t, err, name)
	}
}

// What the instance wrote beside the shipped files stays, so a compiled
// template survives an upgrade.
func TestExtract_Keeps(t *testing.T) {
	t.Setenv(phpfront.Env, "")

	dir := t.TempDir()
	root, err := phpfront.Extract(dir)
	assert.NoError(t, err)

	// A path the shipped tree does not hold: the compiled templates it does
	// ship are replaced on every start, which is how an upgrade applies them.
	own := filepath.Join(root, "theme", "templates_c", "own.tpl")
	assert.NoError(t, os.MkdirAll(filepath.Dir(own), 0o755))
	assert.NoError(t, os.WriteFile(own, []byte("compiled"), 0o644))

	_, err = phpfront.Extract(dir)
	assert.NoError(t, err)
	data, err := os.ReadFile(own)
	assert.NoError(t, err)
	assert.Equal(t, "compiled", string(data))
}

// The checkout override serves a tree in place, so editing a template takes
// effect without a rebuild.
func TestExtract_Checkout(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(phpfront.Env, dir)

	root, err := phpfront.Extract(t.TempDir())
	assert.NoError(t, err)
	assert.Equal(t, dir, root)
}

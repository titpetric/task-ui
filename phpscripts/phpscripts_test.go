package phpscripts_test

import (
	"context"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/phpscript"
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

// The shipped modules, end to end: extract the embedded tree, discover it, run
// every modinfo.php, and render a page. What this covers that the phpscript
// package's own tests cannot is the real module, with composer's autoloader
// resolving MiniTPL inside the filesystem boundary and a template compiling
// into the module's own templates_c.
func TestRefinement(t *testing.T) {
	t.Parallel()
	root := filepath.Join(t.TempDir(), "modules")
	assert.NoError(t, phpscripts.Extract(root))

	var reported []error
	opts := phpscript.NewOptions(root)
	opts.OnError = func(err error) { reported = append(reported, err) }
	m := phpscript.New(opts)
	assert.NoError(t, m.Start(context.Background()))
	assert.Len(t, reported, 0, "every shipped module registers")

	module, ok := m.Module("titpetric/refinement")
	assert.True(t, ok)
	assert.Equal(t, "Refinement", module.Title)
	assert.Equal(t, "flow", module.Icon)
	assert.Equal(t, "/m/titpetric/refinement", module.Href())

	r := httptest.NewRequest(http.MethodGet, "/m/titpetric/refinement", nil)
	out, err := m.Render(context.Background(), httptest.NewRecorder(), r)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, out.Status)
	assert.False(t, out.File)

	body := string(out.Body)
	assert.Contains(t, body, `<section class="refinement">`, "MiniTPL rendered the template")
	assert.Contains(t, body, `data-api="/m/titpetric/refinement/threads"`,
		"the module built its own link from the base task-ui gave it")
	assert.Contains(t, body, "Start a thread")

	// The compiled template is the one the module ships, under its own root.
	assert.FileExists(t, filepath.Join(root, "titpetric", "refinement",
		"templates_c", "templates", "thread.tpl"))
}

// A module's public tree is served as bytes, with no layout around it, so its
// own JavaScript and CSS load.
func TestRefinementAssets(t *testing.T) {
	t.Parallel()
	root := filepath.Join(t.TempDir(), "modules")
	assert.NoError(t, phpscripts.Extract(root))

	m := phpscript.New(phpscript.NewOptions(root))
	assert.NoError(t, m.Start(context.Background()))

	public := filepath.Join(root, "titpetric", "refinement", "public", "assets")
	entries, err := os.ReadDir(public)
	assert.NoError(t, err)
	assert.Greater(t, len(entries), 0, "the module ships at least one asset")

	for _, entry := range entries {
		r := httptest.NewRequest(http.MethodGet, "/m/titpetric/refinement/assets/"+entry.Name(), nil)
		out, err := m.Render(context.Background(), httptest.NewRecorder(), r)
		assert.NoError(t, err, entry.Name())
		assert.True(t, out.File, entry.Name())
		assert.NotEqual(t, "", out.Header.Get("Content-Type"), entry.Name())
	}
}

//go:build integration

package frontend_test

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/frontend"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
	"github.com/titpetric/task-ui/phpscript"
	"github.com/titpetric/task-ui/storage"
)

// withModule writes a module under the fixture's module root, starts a
// manager over it and rebuilds the handler, so the pages serve it.
//
// index is the whole of the module's index.php. files are extra paths
// under the module, for its public tree.
func withModule(t *testing.T, f *fixture, name, index string, files map[string]string) {
	t.Helper()
	dir := filepath.Join(f.modules, filepath.FromSlash(name))
	write := func(rel, body string) {
		path := filepath.Join(dir, filepath.FromSlash(rel))
		assert.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		assert.NoError(t, os.WriteFile(path, []byte(body), 0o644))
	}
	write(phpscript.ModInfo, "<?php\n// @startup\n$m = new Task\\Module;\n"+
		"$m->setName(\""+name+"\");\n$m->setTitle(\"The "+name+"\");\n$m->setIcon(\"flow\");\nTask::register($m);\n")
	write(phpscript.Index, index)
	for rel, body := range files {
		write(rel, body)
	}

	opts := phpscript.NewOptions(f.modules)
	opts.Registry = phpscript.NewStore(storage.NewModuleStorage(f.db), f.opts.Sidebar)
	manager := phpscript.New(opts)
	assert.NoError(t, manager.Start(context.Background()))
	f.opts.Modules = manager

	h, err := frontend.New(f.opts)
	assert.NoError(t, err)
	f.server.Config.Handler = h.Routes()
}

// A module's page is the instance's layout with the module's markup as
// its body: no iframe, the request reaching the PHP unrewritten, and the
// render time in the footer.
func TestModule(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	withModule(t, f, "titpetric/refinement", `<?php
echo "<section class=\"refinement\"><h3>A thread</h3></section>";
echo "<p>uri=", $_SERVER["REQUEST_URI"], "</p>";
echo "<p>base=", $_SERVER["TASK_UI_MODULE_BASE"], "</p>";
`, nil)

	body := f.page(t, asAdmin, "/m/titpetric/refinement/threads/01J?q=one")
	assert.Contains(t, body, `<section class="refinement"><h3>A thread</h3></section>`,
		"the module's markup is drawn, not printed")
	assert.Contains(t, body, "uri=/m/titpetric/refinement/threads/01J?q=one",
		"the PHP reads the whole path, prefix and query included")
	assert.Contains(t, body, "base=/m/titpetric/refinement")
	assert.Contains(t, body, "by phpscript.")
	assert.NotContains(t, body, "<iframe")
	assert.Contains(t, body, `<nav class="menu">`, "the layout is the instance's")

	// The module's row is in the sidebar.
	assert.Contains(t, body, `href="/m/titpetric/refinement"`)
}

// A file under the module's public tree is served as bytes, with its
// content type and no layout, which is how a module's own script loads.
func TestModuleAssets(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	withModule(t, f, "titpetric/refinement", "<?php echo 'the page';", map[string]string{
		"public/assets/refinement.js":  "window.refinement = 1;\n",
		"public/assets/refinement.css": ".refinement { gap: 1rem; }\n",
		"secret.txt":                   "not served",
	})

	res := f.get(t, asAdmin, "/m/titpetric/refinement/assets/refinement.js")
	assert.Equal(t, http.StatusOK, res.Status)
	assert.Equal(t, "window.refinement = 1;\n", res.Body)
	assert.Contains(t, res.Headers.Get("Content-Type"), "javascript")
	assert.NotContains(t, res.Body, "<nav", "a file carries no layout")

	res = f.get(t, asAdmin, "/m/titpetric/refinement/assets/refinement.css")
	assert.Equal(t, http.StatusOK, res.Status)
	assert.Contains(t, res.Headers.Get("Content-Type"), "css")

	// A path outside public reaches index.php, never the file beside it.
	body := f.page(t, asAdmin, "/m/titpetric/refinement/secret.txt")
	assert.NotContains(t, body, "not served")
	assert.Contains(t, body, "the page")
}

// A path naming no module is the not found page, and a module's own 404
// reaches the browser as one.
func TestModuleMissing(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	withModule(t, f, "titpetric/refinement", `<?php
if (strpos($_SERVER["REQUEST_URI"], "/threads/gone") !== false) {
	http_response_code(404);
	echo "No thread there.";
} else {
	echo "the page";
}
`, nil)

	for _, path := range []string{"/m/acme/nothing", "/m/titpetric", "/m"} {
		assert.Equal(t, http.StatusNotFound, f.get(t, asAdmin, path).Status, path)
	}

	res := f.get(t, asAdmin, "/m/titpetric/refinement/threads/gone")
	assert.Equal(t, http.StatusNotFound, res.Status, "the module's own status reaches the browser")
	assert.Contains(t, res.Body, "No thread there.")
}

// Opening a module is the module rule: a viewer without it is refused,
// and the row is out of their sidebar.
func TestModuleRefuses(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	withModule(t, f, "titpetric/refinement", "<?php echo 'the page';", nil)
	ctx := context.Background()

	assert.Equal(t, http.StatusForbidden, f.get(t, asViewer, "/m/titpetric/refinement").Status)
	assert.NotContains(t, f.page(t, asViewer, "/"), `href="/m/titpetric/refinement"`)

	// The rule alone opens it.
	assert.NoError(t, f.opts.Users.Policies().Put(ctx, permission.KindUser, asViewer, "module", permission.Allow))
	res := f.get(t, asViewer, "/m/titpetric/refinement")
	assert.Equal(t, http.StatusOK, res.Status)
	assert.Contains(t, res.Body, "the page")
	assert.Contains(t, res.Body, `href="/m/titpetric/refinement"`)
}

// A module that throws answers 500 and names itself, rather than a blank
// page or a stack trace.
func TestModuleThrows(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	withModule(t, f, "titpetric/thrower", `<?php throw new Exception("the thread is gone");`, nil)

	res := f.get(t, asAdmin, "/m/titpetric/thrower")
	assert.Equal(t, http.StatusInternalServerError, res.Status)
	assert.Contains(t, res.Body, "titpetric/thrower")
}

// A cross origin post to a module is refused: a module is the one part
// of the front end that takes one, so it carries the same check /api has.
func TestModulePostOrigin(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	withModule(t, f, "titpetric/refinement", `<?php echo "method=", $_SERVER["REQUEST_METHOD"];`, nil)

	res := f.send(t, asAdmin, "/m/titpetric/refinement/threads", "", "")
	assert.Equal(t, http.StatusOK, res.Status, "a post with no origin is a non-browser client")
	assert.Contains(t, res.Body, "method=POST")

	res = f.send(t, asAdmin, "/m/titpetric/refinement/threads", "", "https://evil.example")
	assert.Equal(t, http.StatusForbidden, res.Status)
	assert.Contains(t, res.Body, "cross origin")
}

// An instance with no modules answers every module path with the not
// found page rather than a panic, and the sidebar holds the views alone.
func TestModuleNone(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	assert.Equal(t, http.StatusNotFound, f.get(t, asAdmin, "/m/titpetric/refinement").Status)

	rows, err := f.opts.Sidebar.List(context.Background())
	assert.NoError(t, err)
	for _, row := range rows {
		assert.False(t, row.IsModule(), row.Href)
	}
	assert.Len(t, rows, len(model.SidebarDefaults()))
}

//go:build integration

package api_test

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/permission"
	"github.com/titpetric/task-ui/phpscript"
	"github.com/titpetric/task-ui/storage"
)

// withModules writes the named modules under a root of their own, starts
// a manager over it and rebuilds the routes. It returns the root, so a
// test can add a module and reload.
func withModules(t *testing.T, f *fixture, names ...string) string {
	t.Helper()
	root := t.TempDir()
	for _, name := range names {
		writeModule(t, root, name)
	}
	opts := phpscript.NewOptions(root)
	opts.Registry = phpscript.NewStore(storage.NewModuleStorage(f.db), f.opts.Sidebar)
	manager := phpscript.New(opts)
	assert.NoError(t, manager.Start(context.Background()))
	f.opts.Modules = manager
	f.reload(t)
	return root
}

// writeModule puts one module under root: a modinfo.php that registers
// it and an index.php that renders nothing in particular.
func writeModule(t *testing.T, root, name string) {
	t.Helper()
	dir := filepath.Join(root, filepath.FromSlash(name))
	assert.NoError(t, os.MkdirAll(dir, 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, phpscript.ModInfo), []byte(
		"<?php\n// @startup\n$m = new Task\\Module;\n$m->setName(\""+name+"\");\n"+
			"$m->setTitle(\"The "+name+"\");\n$m->setIcon(\"flow\");\n$m->setVersion(\"1.0\");\nTask::register($m);\n"), 0o644))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, phpscript.Index), []byte("<?php echo 'a page';"), 0o644))
}

// The modules list by name, with the link each one opens at.
func TestModules(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	withModules(t, f, "titpetric/refinement", "acme/notes")

	list := f.call(t, asAdmin, http.MethodGet, "/modules", nil)
	assert.Equal(t, http.StatusOK, list.Status)

	rows := list.list("modules")
	assert.Len(t, rows, 2)
	first, _ := rows[0].(map[string]any)
	assert.Equal(t, "acme/notes", first["name"], "the modules come back by name")
	assert.Equal(t, "/m/acme/notes", first["href"])
	assert.Equal(t, "The acme/notes", first["title"])
	assert.Equal(t, "1.0", first["version"])
	assert.Nil(t, first["root"], "the directory on the server's disk is not answered")
}

// A reload reads the roots again, so a module added to a checkout
// appears without a restart and one whose directory is gone goes.
func TestModulesReload(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	root := withModules(t, f, "titpetric/refinement")

	writeModule(t, root, "acme/notes")
	reloaded := f.call(t, asAdmin, http.MethodPost, "/modules/reload", nil)
	assert.Equal(t, http.StatusOK, reloaded.Status)
	assert.Len(t, reloaded.list("modules"), 2)

	assert.NoError(t, os.RemoveAll(filepath.Join(root, "acme")))
	reloaded = f.call(t, asAdmin, http.MethodPost, "/modules/reload", nil)
	assert.Equal(t, http.StatusOK, reloaded.Status)
	assert.Len(t, reloaded.list("modules"), 1)

	// The sidebar row went with it.
	rows, err := f.opts.Sidebar.List(context.Background())
	assert.NoError(t, err)
	for _, row := range rows {
		assert.NotEqual(t, "/m/acme/notes", row.Href)
	}
}

// Listing is the module rule and reloading is an admin's.
func TestModulesRefuses(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	withModules(t, f, "titpetric/refinement")
	ctx := context.Background()

	assert.Equal(t, http.StatusForbidden, f.call(t, asViewer, http.MethodGet, "/modules", nil).Status)
	assert.NoError(t, f.opts.Users.Policies().Put(ctx, permission.KindUser, asViewer, "module", permission.Allow))
	assert.Equal(t, http.StatusOK, f.call(t, asViewer, http.MethodGet, "/modules", nil).Status)

	for _, user := range []string{asOps, asViewer} {
		res := f.call(t, user, http.MethodPost, "/modules/reload", nil)
		assert.Equal(t, http.StatusForbidden, res.Status, user)
	}
}

// An instance with no modules answers an empty list, and a reload says
// it serves none.
func TestModulesNone(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	list := f.call(t, asAdmin, http.MethodGet, "/modules", nil)
	assert.Equal(t, http.StatusOK, list.Status)
	assert.Len(t, list.list("modules"), 0)

	res := f.call(t, asAdmin, http.MethodPost, "/modules/reload", nil)
	assert.Equal(t, http.StatusNotFound, res.Status)
	assert.Contains(t, res.str("error"), "serves no modules")
}

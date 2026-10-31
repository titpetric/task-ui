//go:build integration

package phpscript_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/internal/testdb"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/phpscript"
	"github.com/titpetric/task-ui/storage"
	_ "modernc.org/sqlite"
)

// newStore opens a database of its own and returns the Registry over it,
// with the two storages for a test to read.
func newStore(t *testing.T) (*phpscript.Store, *storage.ModuleStorage, *storage.SidebarStorage) {
	t.Helper()
	db := testdb.New(t)

	modules, sidebar := storage.NewModuleStorage(db), storage.NewSidebarStorage(db)
	return phpscript.NewStore(modules, sidebar), modules, sidebar
}

// A module registers into both tables and the sidebar row links it.
func TestStore(t *testing.T) {
	t.Parallel()
	store, modules, sidebar := newStore(t)
	ctx := context.Background()

	m := &phpscript.Module{
		Vendor: "titpetric", Plugin: "refinement", Root: "/data/modules/titpetric/refinement",
		Title: "Refinement", Description: "Refine a design", Icon: "flow", Version: "0.1.0",
	}
	assert.NoError(t, store.Register(ctx, m))

	row, err := modules.Get(ctx, "titpetric/refinement")
	assert.NoError(t, err)
	assert.Equal(t, "titpetric", row.Vendor)
	assert.Equal(t, "refinement", row.Plugin)
	assert.Equal(t, "Refinement", row.Title)
	assert.Equal(t, "flow", row.Icon)
	assert.Equal(t, "0.1.0", row.Version)

	item, err := sidebar.Get(ctx, "/m/titpetric/refinement")
	assert.NoError(t, err)
	assert.Equal(t, model.SidebarModule, item.Kind)
	assert.Equal(t, "Refinement", item.Title)
	assert.Equal(t, "flow", item.Icon)
	assert.Equal(t, phpscript.ModuleRule, item.Rule)
	assert.False(t, item.Hidden)

	// A module that named no title reads as its plugin, in the table and
	// in the sidebar alike.
	plain := &phpscript.Module{Vendor: "titpetric", Plugin: "mysql", Root: "/x"}
	assert.NoError(t, store.Register(ctx, plain))
	stored, err := modules.Get(ctx, "titpetric/mysql")
	assert.NoError(t, err)
	assert.Equal(t, "mysql", stored.Title)
}

// A module with no name or no root does not reach the tables.
func TestStoreRefuses(t *testing.T) {
	t.Parallel()
	store, modules, _ := newStore(t)
	ctx := context.Background()

	assert.Error(t, store.Register(ctx, &phpscript.Module{Root: "/x"}))
	assert.Error(t, store.Register(ctx, &phpscript.Module{Vendor: "a", Plugin: "b"}))

	rows, err := modules.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, rows, 0)
}

// Forget drops the rows of a module a scan no longer found, and leaves
// the views of the sidebar alone.
func TestStoreForget(t *testing.T) {
	t.Parallel()
	store, modules, sidebar := newStore(t)
	ctx := context.Background()
	assert.NoError(t, sidebar.EnsureDefaults(ctx, model.SidebarDefaults()))

	for _, name := range []string{"a/one", "a/two"} {
		vendor, plugin, err := phpscript.ParseName(name)
		assert.NoError(t, err)
		assert.NoError(t, store.Register(ctx, &phpscript.Module{Vendor: vendor, Plugin: plugin, Root: "/x/" + name}))
	}

	went, err := store.Forget(ctx, []string{"a/one"})
	assert.NoError(t, err)
	assert.Equal(t, 2, went, "the module row and its sidebar row both go")

	rows, err := modules.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, rows, 1)

	items, err := sidebar.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, items, 9, "the eight views stay and one module is left")

	_, err = sidebar.Get(ctx, "/settings")
	assert.NoError(t, err)
}

// The Manager over the real tables: Start registers what it found, and a
// reload after a module is gone takes its rows with it.
func TestStoreThroughManager(t *testing.T) {
	t.Parallel()
	store, modules, sidebar := newStore(t)
	ctx := context.Background()
	assert.NoError(t, sidebar.EnsureDefaults(ctx, model.SidebarDefaults()))

	root := t.TempDir()
	module(t, root, "titpetric/refinement", "<?php echo 'one';")
	module(t, root, "acme/notes", "<?php echo 'two';")

	opts := phpscript.NewOptions(root)
	opts.Registry = store
	m := phpscript.New(opts)
	assert.NoError(t, m.Start(ctx))

	rows, err := modules.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, []string{"acme/notes", "titpetric/refinement"}, []string{rows[0].Name, rows[1].Name})

	items, err := sidebar.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, items, 10, "eight views and a row per module")

	assert.NoError(t, os.RemoveAll(filepath.Join(root, "acme")))
	assert.NoError(t, m.Reload(ctx))

	rows, err = modules.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, rows, 1)
	assert.Equal(t, "titpetric/refinement", rows[0].Name)

	items, err = sidebar.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, items, 9)
}

//go:build integration

package phpfront_test

import (
	"context"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/phpfront"
	"github.com/titpetric/task-ui/storage"
)

// storeDB opens a private database with the schema applied.
func storeDB(t *testing.T) storage.DB {
	t.Helper()
	db, err := storage.OpenAndMigrate(context.Background(), ":memory:", app.Migrations())
	assert.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// A module registers in the module table, and a scan that no longer finds it
// drops its row and keeps the rest.
func TestStore(t *testing.T) {
	t.Parallel()
	db := storeDB(t)
	ctx := context.Background()
	modules := storage.NewModuleStorage(db)
	store := phpfront.NewStore(modules)

	module := &phpfront.Module{Vendor: "taskui", Plugin: "runs", Root: "/srv/runs", Title: "Runs"}
	assert.NoError(t, store.Register(ctx, module))

	rows, err := modules.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, rows, 1)
	assert.Equal(t, "taskui/runs", rows[0].Name)

	old := &phpfront.Module{Vendor: "old", Plugin: "one", Root: "/srv/old", Title: "Old"}
	assert.NoError(t, store.Register(ctx, old))

	dropped, err := store.Forget(ctx, []*phpfront.Module{module})
	assert.NoError(t, err)
	assert.Equal(t, 1, dropped)

	rows, err = modules.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, rows, 1)
	assert.Equal(t, "taskui/runs", rows[0].Name)
}

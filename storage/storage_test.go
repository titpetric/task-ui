package storage_test

import (
	"context"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite" // the driver storage opens

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/storage"
)

// newDB opens a private in-memory database with the schema applied.
func newDB(t *testing.T) storage.DB {
	t.Helper()
	db, err := storage.OpenAndMigrate(context.Background(), ":memory:")
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func TestOpen(t *testing.T) {
	t.Parallel()
	db, err := storage.Open("")
	assert.NoError(t, err)
	var pool storage.DB = db
	assert.NoError(t, pool.Ping())
	assert.Equal(t, 1, db.Stats().MaxOpenConnections)
	db.Close()

	file, err := storage.Open("file:" + filepath.Join(t.TempDir(), "x.db"))
	assert.NoError(t, err)
	assert.NoError(t, file.Ping())
	file.Close()
}

func TestMigrate(t *testing.T) {
	t.Parallel()
	db, err := storage.Open(":memory:")
	assert.NoError(t, err)
	defer db.Close()
	ctx := context.Background()
	assert.NoError(t, storage.Migrate(ctx, db))
	assert.NoError(t, storage.Migrate(ctx, db), "applying twice is fine")

	var count int
	assert.NoError(t, db.GetContext(ctx, &count, "SELECT COUNT(*) FROM migrations WHERE project = ?", storage.Project))
	assert.Equal(t, 24, count)
}

func TestOpenAndMigrate(t *testing.T) {
	t.Parallel()
	db := newDB(t)
	var name string
	assert.NoError(t, db.GetContext(context.Background(), &name, "SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'run'"))
	assert.Equal(t, "run", name)

	_, err := storage.OpenAndMigrate(context.Background(), "file:/nonexistent/dir/x.db")
	assert.Error(t, err)
}

func TestErrNotFound(t *testing.T) {
	t.Parallel()
	assert.Error(t, storage.ErrNotFound)
	assert.Equal(t, "not found", storage.ErrNotFound.Error())
}

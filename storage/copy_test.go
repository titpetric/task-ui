package storage_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

func TestCopy(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dir := t.TempDir()
	db, err := storage.OpenAndMigrate(ctx, filepath.Join(dir, "task-ui.db"))
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	assert.NoError(t, storage.NewSecretStorage(db).Create(ctx, &model.Secret{Scope: "infra", Name: "TOKEN", Content: "x"}))

	path := filepath.Join(dir, "task-ui.db.pre")
	assert.NoError(t, storage.Copy(ctx, db, path))
	copied, err := storage.Open(path)
	assert.NoError(t, err)
	t.Cleanup(func() { copied.Close() })
	list, err := storage.NewSecretStorage(copied).List(ctx)
	assert.NoError(t, err)
	assert.Len(t, list, 1)

	// A second copy leaves the first one untouched.
	before, err := os.ReadFile(path)
	assert.NoError(t, err)
	assert.ErrorIs(t, storage.Copy(ctx, db, path), os.ErrExist)
	after, err := os.ReadFile(path)
	assert.NoError(t, err)
	assert.Equal(t, before, after)
}

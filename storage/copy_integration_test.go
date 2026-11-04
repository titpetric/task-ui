//go:build integration

package storage_test

import (
	"context"
	"errors"
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
	db, err := storage.OpenAndMigrate(ctx, filepath.Join(dir, "task-ui.db"), migrations)
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

// A path Copy can neither read nor write: a directory that is a file,
// which the stat answers for with something other than "not there", and
// a pool that is closed.
func TestCopy_Errors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dir := t.TempDir()
	file := filepath.Join(dir, "file")
	assert.NoError(t, os.WriteFile(file, []byte("x"), 0o600))

	err := storage.Copy(ctx, newDB(t), filepath.Join(file, "copy.db"))
	assert.Error(t, err)
	assert.False(t, errors.Is(err, os.ErrExist), "a file in the way of the directory is not ErrExist")

	assert.ErrorContains(t, storage.Copy(ctx, closedDB(t), filepath.Join(dir, "copy.db")), errClosed)
}

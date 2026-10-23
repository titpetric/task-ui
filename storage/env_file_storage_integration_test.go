//go:build integration

package storage_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

func TestEnvFileStorage_Create(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	files := storage.NewEnvFileStorage(newDB(t))
	file := &model.EnvFile{Folder: "infra", Path: ".env", Content: "A=1\n"}
	assert.NoError(t, files.Create(ctx, file))
	assert.True(t, model.ValidID(file.ID))
	assert.Error(t, files.Create(ctx, &model.EnvFile{Folder: "infra"}), "no path")
	assert.Error(t, files.Create(ctx, &model.EnvFile{Folder: "infra", Path: ".env"}), "duplicate")
}

func TestEnvFileStorage_Get(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	files := storage.NewEnvFileStorage(newDB(t))
	assert.NoError(t, files.Create(ctx, &model.EnvFile{Folder: "infra", Path: ".env", Content: "A=1\n"}))
	got, err := files.Get(ctx, model.EnvFile{Folder: "infra", Path: ".env"})
	assert.NoError(t, err)
	assert.Equal(t, "A=1\n", got.Content)
	_, err = files.Get(ctx, model.EnvFile{Folder: "infra", Path: "other.env"})
	assert.ErrorIs(t, err, storage.ErrNotFound)
}

func TestEnvFileStorage_List(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	files := storage.NewEnvFileStorage(newDB(t))
	assert.NoError(t, files.Create(ctx, &model.EnvFile{Folder: "infra", Path: "conf/b.env"}))
	assert.NoError(t, files.Create(ctx, &model.EnvFile{Folder: "infra", Path: ".env"}))
	assert.NoError(t, files.Create(ctx, &model.EnvFile{Folder: "web", Path: ".env"}))
	list, err := files.List(ctx, "infra")
	assert.NoError(t, err)
	assert.Len(t, list, 2)
	assert.Equal(t, ".env", list[0].Path)
}

func TestEnvFileStorage_Set(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	files := storage.NewEnvFileStorage(newDB(t))
	first := &model.EnvFile{Folder: "infra", Path: ".env", Content: "A=1\n"}
	assert.NoError(t, files.Set(ctx, first))
	second := &model.EnvFile{Folder: "infra", Path: ".env", Content: "A=2\n"}
	assert.NoError(t, files.Set(ctx, second))
	assert.Equal(t, first.ID, second.ID)
	got, err := files.Get(ctx, *first)
	assert.NoError(t, err)
	assert.Equal(t, "A=2\n", got.Content)
}

func TestEnvFileStorage_Update(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	files := storage.NewEnvFileStorage(newDB(t))
	file := &model.EnvFile{Folder: "infra", Path: ".env", Content: "A=1\n"}
	assert.NoError(t, files.Create(ctx, file))
	file.Content = "B=2\n"
	assert.NoError(t, files.Update(ctx, file))
	got, err := files.Get(ctx, *file)
	assert.NoError(t, err)
	assert.Equal(t, "B=2\n", got.Content)
}

func TestEnvFileStorage_Delete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	files := storage.NewEnvFileStorage(newDB(t))
	file := &model.EnvFile{Folder: "infra", Path: ".env"}
	assert.NoError(t, files.Create(ctx, file))
	assert.NoError(t, files.Delete(ctx, file.ID))
	_, err := files.Get(ctx, *file)
	assert.ErrorIs(t, err, storage.ErrNotFound)
}

func TestEnvFileStorage_All(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	files := storage.NewEnvFileStorage(newDB(t))
	assert.NoError(t, files.Create(ctx, &model.EnvFile{Folder: "web", Path: ".env"}))
	assert.NoError(t, files.Create(ctx, &model.EnvFile{Folder: "acme/api", Path: ".env"}))
	all, err := files.All(ctx)
	assert.NoError(t, err)
	assert.Len(t, all, 2)
	assert.Equal(t, "acme/api", all[0].Folder)
}

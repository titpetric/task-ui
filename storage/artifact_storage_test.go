package storage_test

import (
	"context"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

func TestArtifactStorage_Create(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	artifacts := storage.NewArtifactStorage(newDB(t))

	artifact := &model.Artifact{RunID: "r1", Folder: "web", Job: "build", Name: "site.tar.gz", Path: "r1/site.tar.gz", ByteSize: 12, ContentType: "application/gzip"}
	assert.NoError(t, artifacts.Create(ctx, artifact))
	assert.True(t, model.ValidID(artifact.ID))
	assert.NotNil(t, artifact.CreatedAt)
	assert.Error(t, artifacts.Create(ctx, &model.Artifact{RunID: "r1"}), "no name")
}

func TestArtifactStorage_Get(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	artifacts := storage.NewArtifactStorage(newDB(t))
	artifact := &model.Artifact{RunID: "r1", Folder: "web", Job: "build", Name: "a.txt", Path: "r1/a.txt"}
	assert.NoError(t, artifacts.Create(ctx, artifact))
	got, err := artifacts.Get(ctx, artifact.ID)
	assert.NoError(t, err)
	assert.Equal(t, "a.txt", got.Name)
	_, err = artifacts.Get(ctx, "missing")
	assert.ErrorIs(t, err, storage.ErrNotFound)
}

func TestArtifactStorage_List(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	artifacts := storage.NewArtifactStorage(newDB(t))
	base := time.Now().Add(-time.Hour)
	for i, item := range []model.Artifact{
		{RunID: "r1", Folder: "web", Job: "build", Name: "a", Path: "r1/a"},
		{RunID: "r1", Folder: "web", Job: "build", Name: "b", Path: "r1/b"},
		{RunID: "r2", Folder: "infra", Job: "plan", Name: "c", Path: "r2/c"},
	} {
		created := base.Add(time.Duration(i) * time.Minute)
		item.CreatedAt = &created
		assert.NoError(t, artifacts.Create(ctx, &item))
	}
	all, err := artifacts.List(ctx, storage.ArtifactFilter{})
	assert.NoError(t, err)
	assert.Len(t, all, 3)
	assert.Equal(t, "c", all[0].Name, "newest first")

	byRun, err := artifacts.List(ctx, storage.ArtifactFilter{RunID: "r1"})
	assert.NoError(t, err)
	assert.Len(t, byRun, 2)
	byJob, err := artifacts.List(ctx, storage.ArtifactFilter{Folder: "infra", Job: "plan", Limit: 5})
	assert.NoError(t, err)
	assert.Len(t, byJob, 1)
}

func TestArtifactStorage_Delete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	artifacts := storage.NewArtifactStorage(newDB(t))
	one := &model.Artifact{RunID: "r1", Folder: "web", Job: "build", Name: "a", Path: "r1/a"}
	two := &model.Artifact{RunID: "r1", Folder: "web", Job: "build", Name: "b", Path: "r1/b"}
	assert.NoError(t, artifacts.Create(ctx, one))
	assert.NoError(t, artifacts.Create(ctx, two))

	assert.NoError(t, artifacts.Delete(ctx, one.ID))
	_, err := artifacts.Get(ctx, one.ID)
	assert.ErrorIs(t, err, storage.ErrNotFound)

	assert.NoError(t, artifacts.DeleteByRun(ctx, "r1"))
	left, err := artifacts.List(ctx, storage.ArtifactFilter{})
	assert.NoError(t, err)
	assert.Len(t, left, 0)
}

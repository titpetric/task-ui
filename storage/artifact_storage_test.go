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
	byPath, err := artifacts.List(ctx, storage.ArtifactFilter{Path: "r1/b"})
	assert.NoError(t, err)
	assert.Len(t, byPath, 1)
	assert.Equal(t, "b", byPath[0].Name)
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

// Latest is the newest version of each path in the tree; Squash keeps
// those, now in the new commit, and drops the older ones.
func TestArtifactStorage_LatestAndSquash(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	artifacts := storage.NewArtifactStorage(newDB(t))
	create := func(run, name, commit string) {
		assert.NoError(t, artifacts.Create(ctx, &model.Artifact{RunID: run, Folder: "web", Job: "build", Name: name, Path: name, CommitID: commit}))
	}
	create("r1", "a.txt", "c1")
	create("r1", "b.txt", "c1")
	create("r2", "a.txt", "c2")
	assert.NoError(t, artifacts.Create(ctx, &model.Artifact{RunID: "r0", Folder: "web", Job: "old", Name: "x.txt", Path: "r0/x.txt"}))

	latest, err := artifacts.Latest(ctx)
	assert.NoError(t, err)
	assert.Len(t, latest, 2)
	assert.Equal(t, "a.txt", latest[0].Path)
	assert.Equal(t, "r2", latest[0].RunID)
	assert.Equal(t, "b.txt", latest[1].Path)

	assert.NoError(t, artifacts.Squash(ctx, "c3"))
	all, err := artifacts.List(ctx, storage.ArtifactFilter{})
	assert.NoError(t, err)
	assert.Len(t, all, 3, "the older a.txt is gone; the one without a commit stays")
	for _, artifact := range all {
		if artifact.Path != "r0/x.txt" {
			assert.Equal(t, "c3", artifact.CommitID, artifact.Path)
		}
	}

	assert.NoError(t, artifacts.DeleteAll(ctx))
	all, _ = artifacts.List(ctx, storage.ArtifactFilter{})
	assert.Len(t, all, 0)
}

//go:build integration

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

// Rewrite drops the artifacts of dropped commits but the ones it keeps,
// which move to the root, and renames the commits after.
func TestArtifactStorage_Rewrite(t *testing.T) {
	t.Parallel()
	artifacts := storage.NewArtifactStorage(newDB(t))
	ctx := context.Background()
	gone := &model.Artifact{RunID: "r1", Folder: "infra", Job: "build", Name: "a", Path: "a", CommitID: "c1"}
	kept := &model.Artifact{RunID: "r1", Folder: "infra", Job: "build", Name: "b", Path: "b", CommitID: "c1"}
	later := &model.Artifact{RunID: "r2", Folder: "infra", Job: "build", Name: "a", Path: "a", CommitID: "c2"}
	for _, a := range []*model.Artifact{gone, kept, later} {
		assert.NoError(t, artifacts.Create(ctx, a))
	}
	removed, err := artifacts.Rewrite(ctx, "n2", map[string]string{"c2": "n2"}, []string{"c1"}, func(a model.Artifact) bool { return a.ID == kept.ID })
	assert.NoError(t, err)
	assert.Equal(t, 1, removed)
	all, _ := artifacts.List(ctx, storage.ArtifactFilter{})
	assert.Len(t, all, 2)
	for _, a := range all {
		assert.Equal(t, "n2", a.CommitID)
	}
	history, err := artifacts.History(ctx)
	assert.NoError(t, err)
	assert.Len(t, history, 0)
}

// Every method hands back what the driver said.
func TestArtifactStorage_Closed(t *testing.T) {
	t.Parallel()
	artifacts := storage.NewArtifactStorage(closedDB(t))
	ctx := context.Background()

	assert.ErrorContains(t, artifacts.Create(ctx, &model.Artifact{RunID: "r1", Name: "out.txt"}), errClosed)
	_, err := artifacts.Get(ctx, "a1")
	assert.ErrorContains(t, err, errClosed)
	_, err = artifacts.List(ctx, storage.ArtifactFilter{})
	assert.ErrorContains(t, err, errClosed)
	_, err = artifacts.Latest(ctx)
	assert.ErrorContains(t, err, errClosed)
	_, err = artifacts.History(ctx)
	assert.ErrorContains(t, err, errClosed)
	assert.ErrorContains(t, artifacts.Delete(ctx, "a1"), errClosed)
	assert.ErrorContains(t, artifacts.DeleteByRun(ctx, "r1"), errClosed)
	assert.ErrorContains(t, artifacts.DeleteAll(ctx), errClosed)
	assert.ErrorContains(t, artifacts.Squash(ctx, "c1"), errClosed)
	_, err = artifacts.Rewrite(ctx, "c1", nil, nil, nil)
	assert.ErrorContains(t, err, errClosed)
}

// A rewrite stops at the statement that failed and reports it: the read
// of a commit's artifacts, a rename, and the write of a row the keep
// callback cancelled the context before.
func TestArtifactStorage_RewriteErrors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	gone := storage.NewArtifactStorage(brokenDB(t, model.ArtifactTable))
	_, err := gone.Rewrite(ctx, "root", nil, []string{"c1"}, nil)
	assert.ErrorContains(t, err, errNoTable)
	_, err = gone.Rewrite(ctx, "root", map[string]string{"c1": "c2"}, nil, nil)
	assert.ErrorContains(t, err, errNoTable)

	artifacts := storage.NewArtifactStorage(newDB(t))
	assert.NoError(t, artifacts.Create(ctx, &model.Artifact{UserName: "admin", CommitID: "c1", Name: "out.txt", Path: "out.txt"}))
	stopping, cancel := context.WithCancel(ctx)
	_, err = artifacts.Rewrite(stopping, "root", nil, []string{"c1"}, func(model.Artifact) bool {
		cancel()
		return true
	})
	assert.ErrorContains(t, err, "context canceled")
	left, err := artifacts.Latest(ctx)
	assert.NoError(t, err)
	assert.Len(t, left, 1, "the transaction rolled back")
}

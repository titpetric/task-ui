package storage_test

import (
	"context"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

func TestRunStorage_Create(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	runs := storage.NewRunStorage(newDB(t))

	run := &model.Run{Folder: "infra", Job: "deploy", UserName: "ops", TermCols: 100, TermRows: 30}
	assert.NoError(t, runs.Create(ctx, run))
	assert.True(t, model.ValidID(run.ID))
	assert.Equal(t, model.StateRunning, run.State)
	assert.NotNil(t, run.StartedAt)
	assert.Nil(t, run.FinishedAt)

	assert.Error(t, runs.Create(ctx, &model.Run{Folder: "infra"}), "no job")
}

func TestRunStorage_Get(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	runs := storage.NewRunStorage(newDB(t))
	run := &model.Run{Folder: "infra", Job: "deploy", Forced: 1}
	assert.NoError(t, runs.Create(ctx, run))

	got, err := runs.Get(ctx, run.ID)
	assert.NoError(t, err)
	assert.Equal(t, "infra/deploy", got.Key())
	assert.Equal(t, int64(1), got.Forced)
	assert.NotNil(t, got.StartedAt)

	_, err = runs.Get(ctx, "missing")
	assert.ErrorIs(t, err, storage.ErrNotFound)
}

func TestRunStorage_List(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	runs := storage.NewRunStorage(newDB(t))
	base := time.Now().Add(-time.Hour)
	for i, job := range []string{"deploy", "plan", "deploy"} {
		started := base.Add(time.Duration(i) * time.Minute)
		assert.NoError(t, runs.Create(ctx, &model.Run{Folder: "infra", Job: job, UserID: "u1", StartedAt: &started}))
	}
	assert.NoError(t, runs.Create(ctx, &model.Run{Folder: "web", Job: "build", UserID: "u2"}))

	all, err := runs.List(ctx, storage.RunFilter{})
	assert.NoError(t, err)
	assert.Len(t, all, 4)
	assert.Equal(t, "web", all[0].Folder, "newest first")

	deploys, err := runs.List(ctx, storage.RunFilter{Folder: "infra", Job: "deploy"})
	assert.NoError(t, err)
	assert.Len(t, deploys, 2)

	limited, err := runs.List(ctx, storage.RunFilter{Limit: 1, UserID: "u1"})
	assert.NoError(t, err)
	assert.Len(t, limited, 1)
	assert.Equal(t, "u1", limited[0].UserID)

	none, err := runs.List(ctx, storage.RunFilter{State: model.StateOK})
	assert.NoError(t, err)
	assert.Len(t, none, 0)
}

func TestRunStorage_Active(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	runs := storage.NewRunStorage(newDB(t))
	running := &model.Run{Folder: "infra", Job: "deploy"}
	assert.NoError(t, runs.Create(ctx, running))
	done := &model.Run{Folder: "infra", Job: "plan"}
	assert.NoError(t, runs.Create(ctx, done))
	done.State = model.StateOK
	assert.NoError(t, runs.Finish(ctx, done))

	active, err := runs.Active(ctx)
	assert.NoError(t, err)
	assert.Len(t, active, 1)
	assert.Equal(t, running.ID, active[0].ID)
}

func TestRunStorage_Last(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	runs := storage.NewRunStorage(newDB(t))
	_, err := runs.Last(ctx, storage.RunFilter{Folder: "infra", Job: "deploy"})
	assert.ErrorIs(t, err, storage.ErrNotFound)

	earlier := time.Now().Add(-time.Hour)
	assert.NoError(t, runs.Create(ctx, &model.Run{Folder: "infra", Job: "deploy", UserName: "old", StartedAt: &earlier}))
	assert.NoError(t, runs.Create(ctx, &model.Run{Folder: "infra", Job: "deploy", UserName: "new"}))
	last, err := runs.Last(ctx, storage.RunFilter{Folder: "infra", Job: "deploy"})
	assert.NoError(t, err)
	assert.Equal(t, "new", last.UserName)
}

func TestRunStorage_Update(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	runs := storage.NewRunStorage(newDB(t))
	run := &model.Run{Folder: "infra", Job: "deploy"}
	assert.NoError(t, runs.Create(ctx, run))
	run.CastPath = "infra/deploy/x.cast"
	assert.NoError(t, runs.Update(ctx, run))
	got, err := runs.Get(ctx, run.ID)
	assert.NoError(t, err)
	assert.Equal(t, "infra/deploy/x.cast", got.CastPath)
}

func TestRunStorage_Finish(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	runs := storage.NewRunStorage(newDB(t))
	run := &model.Run{Folder: "infra", Job: "deploy"}
	assert.NoError(t, runs.Create(ctx, run))

	assert.Error(t, runs.Finish(ctx, run), "still running")
	run.State = model.StateFailed
	run.ExitCode = 2
	assert.NoError(t, runs.Finish(ctx, run))
	got, err := runs.Get(ctx, run.ID)
	assert.NoError(t, err)
	assert.True(t, got.Finished())
	assert.Equal(t, int64(2), got.ExitCode)
	assert.NotNil(t, got.FinishedAt)
}

func TestRunStorage_Delete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	runs := storage.NewRunStorage(newDB(t))
	run := &model.Run{Folder: "infra", Job: "deploy"}
	assert.NoError(t, runs.Create(ctx, run))
	assert.NoError(t, runs.Delete(ctx, run.ID))
	_, err := runs.Get(ctx, run.ID)
	assert.ErrorIs(t, err, storage.ErrNotFound)
}

func TestRunStorage_KillActive(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	runs := storage.NewRunStorage(newDB(t))
	assert.NoError(t, runs.Create(ctx, &model.Run{Folder: "infra", Job: "deploy"}))
	assert.NoError(t, runs.Create(ctx, &model.Run{Folder: "infra", Job: "plan"}))

	killed, err := runs.KillActive(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), killed)
	active, err := runs.Active(ctx)
	assert.NoError(t, err)
	assert.Len(t, active, 0)
	list, err := runs.List(ctx, storage.RunFilter{State: model.StateKilled})
	assert.NoError(t, err)
	assert.Len(t, list, 2)
	assert.NotNil(t, list[0].FinishedAt)
}

//go:build integration

package runner_test

import (
	"context"
	"io"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
)

// A request with the ID of a run of the same job runs under that ID:
// the new attempt replaces the row, the recording, the summary and the
// artifacts of the run before. Start refuses another job's run, or an
// ID that is no ID.
func TestManager_StartReuse(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	req := f.shellJob("again", `echo first; echo one > "$TASK_UI_ARTIFACTS/first.txt"`, false)
	first, err := f.manager.Start(ctx, req)
	assert.NoError(t, err)
	collect(t, first.Attach(true))
	<-first.Done()
	assert.NoError(t, f.summaries.Set(ctx, &model.RunSummary{RunID: first.ID(), Summary: "it said first"}))

	req.ID = first.ID()
	req.Job.Command = []string{"sh", "-c", `echo second; echo two > "$TASK_UI_ARTIFACTS/second.txt"; exit 2`}
	second, err := f.manager.Start(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, first.ID(), second.ID())
	_, status := collect(t, second.Attach(true))
	assert.Equal(t, model.StateFailed, status.State)
	<-second.Done()

	run, err := f.runs.Get(ctx, first.ID())
	assert.NoError(t, err)
	assert.Equal(t, model.StateFailed, run.State)
	assert.Equal(t, int64(2), run.ExitCode)
	runs, err := f.runs.List(ctx, storage.RunFilter{Folder: "again"})
	assert.NoError(t, err)
	assert.Len(t, runs, 1, "one row for the two attempts")
	cast, err := f.recordings.Read(run)
	assert.NoError(t, err)
	defer cast.Close()
	recorded, err := io.ReadAll(cast)
	assert.NoError(t, err)
	assert.Contains(t, string(recorded), "second")
	assert.NotContains(t, string(recorded), "first")
	_, err = f.summaries.Get(ctx, first.ID())
	assert.ErrorIs(t, err, storage.ErrNotFound, "the summary of the attempt before goes")
	artifacts, err := f.files.List(ctx, storage.ArtifactFilter{RunID: first.ID()})
	assert.NoError(t, err)
	assert.Len(t, artifacts, 1)
	assert.Equal(t, "second.txt", artifacts[0].Name)
	kept, ok := f.manager.Get(first.ID())
	assert.True(t, ok)
	assert.Equal(t, second, kept, "the session of the ID is the latest attempt")

	other := f.shellJob("other", "true", false)
	other.ID = first.ID()
	_, err = f.manager.Start(ctx, other)
	assert.ErrorIs(t, err, runner.ErrReuse)
	other.ID = "../x"
	_, err = f.manager.Start(ctx, other)
	assert.ErrorIs(t, err, runner.ErrReuse)
	_, ok = f.manager.Active(other.Folder.ID + "/run")
	assert.False(t, ok, "a refused run holds no lock")

	fresh := f.shellJob("fresh", "true", false)
	fresh.ID = model.NewID()
	started, err := f.manager.Start(ctx, fresh)
	assert.NoError(t, err)
	assert.Equal(t, fresh.ID, started.ID(), "an ID with no run takes it")
	<-started.Done()
}

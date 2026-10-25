//go:build integration

package runner_test

import (
	"context"
	"testing"
	"time"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
)

// A job with hours runs inside them, refuses outside them, and runs
// outside them when forced with a reason, which the run keeps.
func TestManager_Gate(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	req := f.shellJob("gated", "echo gated", false)
	req.Folder.Hours = "mon-fri 9-17\nfri 9-13"

	friday := time.Date(2026, time.October, 16, 10, 0, 0, 0, time.Local)
	f.now = friday
	session, err := f.manager.Start(context.Background(), req)
	assert.NoError(t, err)
	_, status := collect(t, session.Attach(false))
	assert.Equal(t, model.StateOK, status.State)
	assert.Equal(t, "", session.Snapshot().Reason)

	f.now = friday.Add(5 * time.Hour)
	_, err = f.manager.Start(context.Background(), req)
	assert.ErrorIs(t, err, runner.ErrOutOfHours)

	req.Force = true
	_, err = f.manager.Start(context.Background(), req)
	assert.ErrorIs(t, err, runner.ErrReasonNeeded)

	req.Reason = "hotfix for the outage"
	session, err = f.manager.Start(context.Background(), req)
	assert.NoError(t, err)
	_, status = collect(t, session.Attach(false))
	assert.Equal(t, model.StateOK, status.State)
	run, err := f.runs.Get(context.Background(), session.ID())
	assert.NoError(t, err)
	assert.Equal(t, "hotfix for the outage", run.Reason)
	assert.Equal(t, int64(1), run.Forced)
}

// The schedule skips the gate, a built-in takes it as any job does, and
// bad hours refuse the run with what is wrong.
func TestManager_GateSchedule(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	req := f.shellJob("gated", "echo gated", false)
	req.Folder.Hours = "sun 0-1"
	f.now = time.Date(2026, time.October, 14, 12, 0, 0, 0, time.Local)

	shell := req
	shell.Job = req.Folder.Job("@shell")
	if shell.Job != nil {
		_, err := f.manager.Start(context.Background(), shell)
		assert.ErrorIs(t, err, runner.ErrOutOfHours)
	}

	_, err := f.manager.Start(context.Background(), req)
	assert.ErrorIs(t, err, runner.ErrOutOfHours)

	req.Scheduled = true
	session, err := f.manager.Start(context.Background(), req)
	assert.NoError(t, err)
	collect(t, session.Attach(false))

	req.Scheduled = false
	req.Folder.Hours = "someday 9-17"
	_, err = f.manager.Start(context.Background(), req)
	assert.ErrorContains(t, err, "invalid hours")
}

// A prompt's run, @claude-prompt in the folder it asks in, takes no
// hours of the folder and gets the folder's secrets, as any job there
// does; its row names the folder, the directory and the job.
func TestManager_PromptJob(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	assert.NoError(t, f.secrets.Set(ctx, &model.Secret{Scope: storage.ScopeAll, Name: "SHARED", Content: "shared"}))
	assert.NoError(t, f.secrets.Set(ctx, &model.Secret{Scope: "asked", Name: "OWN", Content: "own"}))

	req := f.shellJob("asked", "", false)
	req.Folder.Hours = "sun 0-1"
	f.now = time.Date(2026, time.October, 14, 12, 0, 0, 0, time.Local)
	job := req.Folder.PromptJob()
	assert.True(t, job.IsPrompt())
	job.Command = []string{"sh", "-c", `echo "[$SHARED][$OWN] $TASK_UI_JOB"`}
	req.Job = job

	session, err := f.manager.Start(ctx, req)
	assert.NoError(t, err, "out of the folder's hours")
	out, status := collect(t, session.Attach(true))
	assert.Equal(t, model.StateOK, status.State)
	assert.Contains(t, out, "[shared][own] @claude-prompt")
	run := session.Snapshot()
	assert.Equal(t, "asked", run.Folder)
	assert.Equal(t, folder.PromptID, run.Job)
	assert.True(t, run.IsPrompt())
}

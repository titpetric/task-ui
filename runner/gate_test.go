package runner_test

import (
	"context"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
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

// The job's own hours win over the folder's, the schedule skips the gate,
// and bad hours refuse the run with what is wrong.
func TestManager_GateJobAndSchedule(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	req := f.shellJob("gated", "echo gated", false)
	req.Folder.Hours = "sun 0-1"
	req.Job.Hours = "daily 0-24"
	f.now = time.Date(2026, time.October, 14, 12, 0, 0, 0, time.Local)

	session, err := f.manager.Start(context.Background(), req)
	assert.NoError(t, err)
	collect(t, session.Attach(false))

	req.Job.Hours = ""
	_, err = f.manager.Start(context.Background(), req)
	assert.ErrorIs(t, err, runner.ErrOutOfHours)

	req.Scheduled = true
	session, err = f.manager.Start(context.Background(), req)
	assert.NoError(t, err)
	collect(t, session.Attach(false))

	req.Scheduled = false
	req.Folder.Hours = "someday 9-17"
	_, err = f.manager.Start(context.Background(), req)
	assert.ErrorContains(t, err, "invalid hours")
}

package runner_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
)

// A job runs once at a time: a second start finds the run in the way,
// and a forced one kills it and takes its place; the lock goes with
// the run that ended.
func TestManager_Lock(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	slow, err := f.manager.Start(ctx, f.shellJob("lock", "sleep 30", false))
	assert.NoError(t, err)
	in, err := f.manager.Start(ctx, f.shellJob("lock", "true", false))
	assert.ErrorIs(t, err, runner.ErrBusy)
	assert.Equal(t, slow.ID(), in.ID())

	forced := f.shellJob("lock", "true", false)
	forced.Force = true
	next, err := f.manager.Start(ctx, forced)
	assert.NoError(t, err)
	wait(t, slow)
	assert.Equal(t, model.StateKilled, slow.Status().State)
	wait(t, next)
	again, err := f.manager.Start(ctx, f.shellJob("lock", "true", false))
	assert.NoError(t, err)
	wait(t, again)
}

// The runs of a prompt go side by side in one folder, each holding a
// lock of its own, while an ordinary job of the folder still runs once
// at a time.
func TestManager_LockParallel(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	prompt := func() runner.Request {
		req := f.shellJob("prompts", "sleep 30", false)
		req.Job.ID = model.PromptJob
		return req
	}
	first, err := f.manager.Start(ctx, prompt())
	assert.NoError(t, err)
	second, err := f.manager.Start(ctx, prompt())
	assert.NoError(t, err)
	assert.NotEqual(t, first.ID(), second.ID())
	assert.NotEqual(t, first.Key(), second.Key())
	assert.Equal(t, model.StateRunning, first.Status().State)
	assert.Equal(t, model.StateRunning, second.Status().State)

	job, err := f.manager.Start(ctx, f.shellJob("prompts", "sleep 30", false))
	assert.NoError(t, err)
	in, err := f.manager.Start(ctx, f.shellJob("prompts", "true", false))
	assert.ErrorIs(t, err, runner.ErrBusy)
	assert.Equal(t, job.ID(), in.ID())
	for _, session := range []*runner.Session{first, second, job} {
		assert.NoError(t, f.manager.Kill(session.ID()))
		wait(t, session)
	}
}

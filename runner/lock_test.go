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

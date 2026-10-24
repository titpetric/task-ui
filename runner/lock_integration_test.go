//go:build integration

package runner_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
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

// Starts of one job at once: one runs, the others find it in the way,
// with its session once its row exists and none before, and no start
// panics on a start under way. Forced starts in the same window are as
// safe, and once the runs end nothing waits, queued, for a lock nobody
// holds.
func TestManager_LockConcurrent(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	const starts = 8
	var wg sync.WaitGroup
	sessions, errs := make([]*runner.Session, starts), make([]error, starts)
	for i := range starts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sessions[i], errs[i] = f.manager.Start(ctx, f.shellJob("race", "sleep 30", false))
		}()
	}
	wg.Wait()
	started := 0
	var running *runner.Session
	for i := range starts {
		if errs[i] == nil {
			started++
			running = sessions[i]
			continue
		}
		assert.ErrorIs(t, errs[i], runner.ErrBusy)
	}
	assert.Equal(t, 1, started, "one start takes the lock")
	for i := range starts {
		if errs[i] != nil && sessions[i] != nil {
			assert.Equal(t, running.ID(), sessions[i].ID(), "the session in the way is the run that started")
		}
	}

	forced := make([]error, starts)
	for i := range starts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := f.shellJob("race", "sleep 30", false)
			req.Force = true
			_, forced[i] = f.manager.Start(ctx, req)
		}()
	}
	wg.Wait()
	for _, err := range forced {
		if err != nil {
			assert.ErrorIs(t, err, runner.ErrBusy)
		}
	}
	wait(t, running)
	assert.Equal(t, model.StateKilled, running.Status().State)

	current, ok := f.manager.Active("race/run")
	assert.True(t, ok, "a forced start holds the lock")
	assert.NotNil(t, current)
	assert.NoError(t, f.manager.Kill(current.ID()))
	wait(t, current)
	deadline := time.Now().Add(5 * time.Second)
	for len(f.manager.List()) > 0 && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	assert.Len(t, f.manager.List(), 0, "no run holds or waits for the lock")
	runs, err := f.runs.List(ctx, storage.RunFilter{State: model.StateQueued})
	assert.NoError(t, err)
	assert.Len(t, runs, 0, "no run stays queued")
}

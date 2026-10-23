//go:build integration

package runner_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
)

// storageQueued keeps the queued runs of the job q/run.
var storageQueued = storage.RunFilter{Folder: "q", State: model.StateQueued}

// queueJob is a run of the job q/run that appends its name to the order
// file once the gate file is there, as a request that queues.
func (f *fixture) queueJob(name, gate, order string) runner.Request {
	req := f.shellJob("q", `while [ ! -f "`+gate+`" ]; do sleep 0.05; done; echo `+name+` >> "`+order+`"`, false)
	req.Queue = true
	return req
}

// wait waits for a session to settle.
func wait(t *testing.T, session *runner.Session) {
	t.Helper()
	select {
	case <-session.Done():
	case <-time.After(10 * time.Second):
		t.Fatal("the run did not end")
	}
}

// A run of a job that runs waits, queued, and starts once the run before
// it ended, in the order the runs came; a stop takes a queued run out of
// the queue, and the runs behind it move up.
func TestManager_Queue(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	dir := t.TempDir()
	gate, order := filepath.Join(dir, "gate"), filepath.Join(dir, "order")

	first, err := f.manager.Start(ctx, f.queueJob("first", gate, order))
	assert.NoError(t, err)
	second, err := f.manager.Start(ctx, f.queueJob("second", gate, order))
	assert.NoError(t, err)
	dropped, err := f.manager.Start(ctx, f.queueJob("dropped", gate, order))
	assert.NoError(t, err)
	third, err := f.manager.Start(ctx, f.queueJob("third", gate, order))
	assert.NoError(t, err)
	assert.Equal(t, model.StateRunning, first.Status().State)
	for _, session := range []*runner.Session{second, dropped, third} {
		assert.Equal(t, model.StateQueued, session.Status().State)
		row, err := f.runs.Get(ctx, session.ID())
		assert.NoError(t, err)
		assert.Equal(t, model.StateQueued, row.State)
	}
	listed, err := f.runs.List(ctx, storageQueued)
	assert.NoError(t, err)
	assert.Len(t, listed, 3)

	assert.NoError(t, f.manager.Stop(dropped.ID()))
	wait(t, dropped)
	row, err := f.runs.Get(ctx, dropped.ID())
	assert.NoError(t, err)
	assert.Equal(t, model.StateStopped, row.State)
	assert.NotNil(t, row.FinishedAt)

	assert.NoError(t, os.WriteFile(gate, nil, 0o644))
	for _, session := range []*runner.Session{first, second, third} {
		wait(t, session)
		assert.Equal(t, model.StateOK, session.Status().State)
	}
	got, err := os.ReadFile(order)
	assert.NoError(t, err)
	assert.Equal(t, "first\nsecond\nthird\n", string(got))
	row, err = f.runs.Get(ctx, third.ID())
	assert.NoError(t, err)
	assert.Equal(t, model.StateOK, row.State)
	assert.True(t, row.StartedAt.After(*second.Snapshot().StartedAt), "the start is when it left the queue")

	// A request that doesn't queue still finds the job busy.
	busy, err := f.manager.Start(ctx, f.queueJob("x", gate, order))
	assert.NoError(t, err)
	plain := f.shellJob("q", "true", false)
	_, err = f.manager.Start(ctx, plain)
	assert.ErrorIs(t, err, runner.ErrBusy)
	wait(t, busy)
}

// At most MaxQueued runs wait for a job; Start refuses the next.
func TestManager_QueueFull(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	dir := t.TempDir()
	gate, order := filepath.Join(dir, "gate"), filepath.Join(dir, "order")
	var sessions []*runner.Session
	for i := 0; i <= runner.MaxQueued; i++ {
		session, err := f.manager.Start(ctx, f.queueJob("run", gate, order))
		assert.NoError(t, err)
		sessions = append(sessions, session)
	}
	_, err := f.manager.Start(ctx, f.queueJob("over", gate, order))
	assert.True(t, errors.Is(err, runner.ErrQueueFull), err)
	assert.Contains(t, err.Error(), "10 runs of the job wait already")
	listed, err := f.runs.List(ctx, storageQueued)
	assert.NoError(t, err)
	assert.Len(t, listed, runner.MaxQueued)

	assert.NoError(t, os.WriteFile(gate, nil, 0o644))
	for _, session := range sessions {
		wait(t, session)
	}
}

// Shutdown ends the queued runs failed, saying the instance went down
// before they ran.
func TestManager_QueueShutdown(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	dir := t.TempDir()
	gate, order := filepath.Join(dir, "gate"), filepath.Join(dir, "order")
	first, err := f.manager.Start(ctx, f.queueJob("first", gate, order))
	assert.NoError(t, err)
	queued, err := f.manager.Start(ctx, f.queueJob("second", gate, order))
	assert.NoError(t, err)
	f.manager.Shutdown(ctx)
	wait(t, first)
	wait(t, queued)
	row, err := f.runs.Get(ctx, queued.ID())
	assert.NoError(t, err)
	assert.Equal(t, model.StateFailed, row.State)
	assert.Equal(t, model.QueuedLost, row.Message)
	_, err = os.Stat(order)
	assert.True(t, os.IsNotExist(err), "neither run got to write")
}

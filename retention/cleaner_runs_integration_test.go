//go:build integration

package retention_test

import (
	"bytes"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/retention"
	"github.com/titpetric/task-ui/storage"
)

// Recordings older than the rule go, the oldest first past the limit;
// a running run's, and a running workflow's step's, stay; the run rows
// stay; a dry run removes nothing.
func TestCleaner_Recordings(t *testing.T) {
	t.Parallel()
	f := newCleaner(t)
	old := f.recorded(t, 40, 100, model.StateOK)
	running := f.recorded(t, 40, 100, model.StateRunning)
	step := f.recorded(t, 39, 100, model.StateFailed)
	mid := f.recorded(t, 10, 100, model.StateOK)
	fresh := f.recorded(t, 1, 100, model.StateOK)
	flows := f.cleaner.Flows
	flow := &model.WorkflowRun{WorkflowID: "w1", Name: "ship"}
	assert.NoError(t, flows.Create(f.ctx, flow, []model.WorkflowStep{{Folder: "internal/tools", Job: "build"}}))
	steps, _ := flows.Steps(f.ctx, flow.ID)
	steps[0].RunID = step.ID
	assert.NoError(t, flows.UpdateStep(f.ctx, &steps[0]))
	assert.NoError(t, f.cleaner.Save(f.ctx, keep(retention.Recordings, retention.Month)))

	usage, err := f.cleaner.Usage(f.ctx)
	assert.NoError(t, err)
	assert.Equal(t, 5, usage[retention.Recordings].Count)
	assert.Equal(t, int64(500), usage[retention.Recordings].Bytes)
	assert.Equal(t, 5, usage[retention.Runs].Count)
	assert.Equal(t, int64(500), usage[retention.Runs].Bytes, "run history's size is its recordings'")

	var out bytes.Buffer
	assert.NoError(t, f.cleaner.Clean(f.ctx, true, &out))
	assert.Contains(t, out.String(), "Run recordings (keep for 31 days, no size limit): would remove 1 recording, 100 B; 4 recordings, 400 B left")
	assert.True(t, f.exists(old))

	out.Reset()
	assert.NoError(t, f.cleaner.Clean(f.ctx, false, &out))
	assert.Contains(t, out.String(), "Run recordings (keep for 31 days, no size limit): removed 1 recording, 100 B; 4 recordings, 400 B left")
	assert.False(t, f.exists(old))
	assert.True(t, f.exists(running))
	assert.True(t, f.exists(step))
	row, err := f.cleaner.Runs.Get(f.ctx, old.ID)
	assert.NoError(t, err, "the run stays")
	assert.Equal(t, old.CastPath, row.CastPath)

	// A size limit takes the oldest idle ones until the rest, the busy
	// ones among them, fit.
	s := keep()
	s[retention.Recordings] = retention.Rule{Age: retention.Never, Limit: 350}
	assert.NoError(t, f.cleaner.Save(f.ctx, s))
	out.Reset()
	assert.NoError(t, f.cleaner.Clean(f.ctx, false, &out))
	assert.False(t, f.exists(mid))
	assert.True(t, f.exists(fresh))
	assert.True(t, f.exists(running))
	assert.True(t, f.exists(step))
	assert.NotContains(t, out.String(), "Database: vacuumed", "no rows went")
}

// Run history takes a run older than the age with its recording, its
// summary and the threads about it; a running or queued run, a step of a
// running workflow and the cleanup's own run stay, and so do the
// threads of runs that stay.
func TestCleaner_Runs(t *testing.T) {
	t.Parallel()
	f := newCleaner(t)
	old := f.recorded(t, 40, 100, model.StateOK)
	queued := f.recorded(t, 40, 100, model.StateQueued)
	running := f.recorded(t, 40, 100, model.StateRunning)
	step := f.recorded(t, 39, 100, model.StateFailed)
	self := f.recorded(t, 40, 100, model.StateOK)
	fresh := f.recorded(t, 1, 100, model.StateOK)
	f.cleaner.Self = self.ID
	flows := f.cleaner.Flows
	flow := &model.WorkflowRun{WorkflowID: "w1", Name: "ship"}
	assert.NoError(t, flows.Create(f.ctx, flow, []model.WorkflowStep{{Folder: "internal/tools", Job: "build"}}))
	steps, _ := flows.Steps(f.ctx, flow.ID)
	steps[0].RunID = step.ID
	assert.NoError(t, flows.UpdateStep(f.ctx, &steps[0]))
	assert.NoError(t, f.cleaner.Summaries.Set(f.ctx, &model.RunSummary{RunID: old.ID, Summary: "it built"}))
	assert.NoError(t, f.cleaner.Summaries.Set(f.ctx, &model.RunSummary{RunID: fresh.ID, Summary: "it built again"}))
	gone := f.thread(t, 40, old.ID, "")
	stays := f.thread(t, 40, fresh.ID, "")
	assert.NoError(t, f.cleaner.Save(f.ctx, keep(retention.Runs, retention.Month)))

	var out bytes.Buffer
	assert.NoError(t, f.cleaner.Clean(f.ctx, false, &out))
	assert.Contains(t, out.String(), "Run history (keep for 31 days, no size limit): removed 1 run, 100 B; 5 runs, 500 B left")
	assert.Contains(t, out.String(), "Prompt threads (keep for ever): removed 0 threads; 1 thread left", "the thread run history took counts to it")
	assert.Contains(t, out.String(), "Database: vacuumed, ")
	_, err := f.cleaner.Runs.Get(f.ctx, old.ID)
	assert.ErrorIs(t, err, storage.ErrNotFound)
	assert.False(t, f.exists(old))
	_, err = f.cleaner.Summaries.Get(f.ctx, old.ID)
	assert.ErrorIs(t, err, storage.ErrNotFound)
	_, err = f.cleaner.Notes.Get(f.ctx, gone.ID)
	assert.ErrorIs(t, err, storage.ErrNotFound)
	replies, _ := f.cleaner.Notes.Replies(f.ctx, []string{gone.ID})
	assert.Len(t, replies[gone.ID], 0)
	for _, run := range []*model.Run{queued, running, step, self, fresh} {
		_, err := f.cleaner.Runs.Get(f.ctx, run.ID)
		assert.NoError(t, err, run.State)
		assert.True(t, f.exists(run))
	}
	_, err = f.cleaner.Notes.Get(f.ctx, stays.ID)
	assert.NoError(t, err)
	summary, err := f.cleaner.Summaries.Get(f.ctx, fresh.ID)
	assert.NoError(t, err)
	assert.Equal(t, "it built again", summary.Summary)
}

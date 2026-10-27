//go:build integration

package frontend_test

import (
	"context"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// clock is the fixture's pinned time, the day the charts end on.
var clock = time.Date(2026, time.October, 16, 22, 0, 0, 0, time.Local)

// ran records a finished run of a job at a time, lasting a while.
func (f *fixture) ran(t *testing.T, folderID, job, state string, at time.Time, took time.Duration) {
	t.Helper()
	end := at.Add(took)
	run := &model.Run{Folder: folderID, Job: job, UserName: asOps, State: state, StartedAt: &at, FinishedAt: &end}
	assert.NoError(t, f.opts.Runs.Create(context.Background(), run))
}

// The runs page charts the runs of the last thirty days by day and of the
// last 24 hours by hour, across jobs and without a duration line; a
// folder the user may not see is not counted.
func TestActivity_Runs(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.ran(t, "infra", "plan", model.StateOK, clock.Add(-2*time.Hour), 2*time.Second)
	f.ran(t, "infra", "plan", model.StateFailed, clock.Add(-3*time.Hour), 3*time.Second)
	f.ran(t, "private", "deploy", model.StateOK, clock.Add(-26*time.Hour), time.Second)
	f.ran(t, "infra", "plan", model.StateOK, clock.AddDate(0, 0, -40), time.Second) // before the window

	admin := f.page(t, asAdmin, "/runs")
	assert.Contains(t, admin, `<div class="activity-bars" style="--n: 30"`)
	assert.Contains(t, admin, "3 runs in 30 days")
	assert.Contains(t, admin, `<span class="bad">1 failed</span>`)
	assert.Contains(t, admin, `data-tip="2 runs, Fri 2026-10-16 · 1 failed"`)
	assert.Contains(t, admin, `data-tip="1 run, Thu 2026-10-15"`)
	// The last 24 hours hold the 16th's two runs; the 15th's is older.
	assert.Contains(t, admin, `data-tip="1 run, Fri 20:00–21:00"`)
	assert.Contains(t, admin, `data-tip="1 run, Fri 19:00–20:00 · 1 failed"`)
	// Each chart heads its own window's finished runs.
	assert.Contains(t, admin, `<span>Last 30 days</span>`)
	assert.Contains(t, admin, `min/avg/max/total: 1.0s 2.0s 3.0s 6.0s</span>`)
	assert.Contains(t, admin, `<span>Last 24 hours</span>`)
	assert.Contains(t, admin, `min/avg/max/total: 2.0s 2.5s 3.0s 5.0s</span>`)
	// The durations of the days and the hours as lines; no line of the
	// latest runs, which would mix jobs.
	assert.Contains(t, admin, `class="activity-line durations"`)
	assert.NotContains(t, admin, `<div class="activity-line">`)

	limited := f.page(t, asLimited, "/runs")
	assert.Contains(t, limited, "2 runs in 30 days")
	assert.Contains(t, limited, `data-tip="no runs, Thu 2026-10-15"`)

	filtered := f.page(t, asAdmin, "/runs?job=deploy")
	assert.Contains(t, filtered, "1 run in 30 days")
}

// A job's page charts its own runs, the duration line with them.
func TestActivity_Job(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.ran(t, "infra", "plan", model.StateOK, clock.Add(-2*time.Hour), 2*time.Second)
	f.ran(t, "infra", "plan", model.StateFailed, clock.Add(-time.Hour), 4*time.Second)
	f.ran(t, "infra", "slow", model.StateOK, clock.Add(-time.Hour), time.Minute)

	body := f.page(t, asOps, "/f/internal/infra/plan")
	assert.Contains(t, body, "2 runs in 30 days")
	assert.Contains(t, body, "median 3.0s")
	assert.Contains(t, body, `class="activity-line"`)
	assert.Contains(t, body, `<line class="bad"`)
	assert.Contains(t, body, "4.0s, Fri 2026-10-16 21:00 · failed")
	assert.Contains(t, body, `class="activity-line durations"`)
	assert.Contains(t, body, `data-tip="avg 3.0s · min 2.0s · max 4.0s over 2 runs, Fri 2026-10-16"`)

	quiet := f.page(t, asOps, "/f/internal/infra/many")
	assert.NotContains(t, quiet, `class="activity"`)
}

// A workflow's page charts its runs and how long they took.
func TestActivity_Workflow(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	wf := f.workflow(t, asOps, "release", "infra", "plan")
	for i, state := range []string{model.StateOK, model.StateFailed} {
		run := &model.WorkflowRun{WorkflowID: wf.ID, Name: wf.Name, UserName: asOps}
		assert.NoError(t, f.opts.WorkflowRuns.Create(ctx, run, nil))
		start := clock.Add(-time.Duration(i+1) * time.Hour)
		end := start.Add(time.Duration(i+1) * 10 * time.Second)
		run.State, run.StartedAt, run.FinishedAt = state, &start, &end
		assert.NoError(t, f.opts.WorkflowRuns.Update(ctx, run))
	}

	body := f.page(t, asOps, "/workflows/"+wf.ID)
	assert.Contains(t, body, "2 runs in 30 days")
	assert.Contains(t, body, "median 15.0s")
	assert.Contains(t, body, `data-tip="2 runs, Fri 2026-10-16 · 1 failed"`)
	assert.Contains(t, body, `class="activity-line"`)

	empty := f.workflow(t, asOps, "never", "infra", "plan")
	assert.NotContains(t, f.page(t, asOps, "/workflows/"+empty.ID), `class="activity"`)
}

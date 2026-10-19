package view_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// Recents interleaves job runs and workflow runs by when they started,
// newest first, and keeps the first n.
func TestRecents(t *testing.T) {
	t.Parallel()
	at := func(m int) *time.Time { v := time.Date(2026, 10, 19, 10, m, 0, 0, time.UTC); return &v }
	runs := []view.Run{
		{Run: model.Run{ID: "r1", StartedAt: at(1)}},
		{Run: model.Run{ID: "r3", StartedAt: at(3)}},
	}
	flows := []model.WorkflowRun{{ID: "w2", StartedAt: at(2)}, {ID: "w4", StartedAt: at(4)}}

	got := view.Recents(runs, flows, 3)
	assert.Len(t, got, 3)
	assert.Equal(t, "w4", got[0].Flow.ID)
	assert.Equal(t, "r3", got[1].Run.ID)
	assert.Equal(t, "w2", got[2].Flow.ID)
}

// An entry starts and lasts as its run, of a job or of a workflow, and
// carries the initials of who started it.
func TestRecent_StartedAt(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, 10, 19, 10, 0, 0, 0, time.UTC)
	assert.Equal(t, start, view.Recent{Run: &view.Run{Run: model.Run{StartedAt: &start}}}.StartedAt())
	assert.Equal(t, start, view.Recent{Flow: &model.WorkflowRun{StartedAt: &start}}.StartedAt())
	assert.True(t, view.Recent{Flow: &model.WorkflowRun{}}.StartedAt().IsZero())
	assert.True(t, view.Recent{}.StartedAt().IsZero())
}

func TestRecent_Elapsed(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, 10, 19, 10, 0, 0, 0, time.UTC)
	end := start.Add(90 * time.Second)
	assert.Equal(t, 90*time.Second, view.Recent{Run: &view.Run{Run: model.Run{StartedAt: &start, FinishedAt: &end}}}.Elapsed())
	assert.Equal(t, 90*time.Second, view.Recent{Flow: &model.WorkflowRun{StartedAt: &start, FinishedAt: &end}}.Elapsed())
	running := time.Now().Add(-time.Minute)
	assert.True(t, view.Recent{Flow: &model.WorkflowRun{StartedAt: &running}}.Elapsed() >= time.Minute)
	assert.Equal(t, time.Duration(0), view.Recent{Flow: &model.WorkflowRun{}}.Elapsed())
	assert.Equal(t, time.Duration(0), view.Recent{}.Elapsed())
}

func TestRecent_Initials(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "BL", view.Recent{Run: &view.Run{Run: model.Run{UserName: "black"}}}.Initials())
	assert.Equal(t, "TP", view.Recent{Flow: &model.WorkflowRun{UserName: "tit petric"}}.Initials())
}

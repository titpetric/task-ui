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

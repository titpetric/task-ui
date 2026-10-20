package view_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// A window's stats count its finished runs only.
func TestActivityStats(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)
	a := view.NewActivity([]view.ActivityRun{
		{StartedAt: now.Add(-time.Hour), Elapsed: time.Second, State: model.StateOK},
		{StartedAt: now.Add(-2 * time.Hour), Elapsed: 3 * time.Second, State: model.StateFailed},
		{StartedAt: now.Add(-3 * time.Hour), State: model.StateRunning},
	}, now, 1)
	var stats view.ActivityStats = a.HourStats
	assert.Equal(t, view.ActivityStats{Count: 2, Fastest: time.Second, Mean: 2 * time.Second, Slowest: 3 * time.Second, Spent: 4 * time.Second}, stats)
}

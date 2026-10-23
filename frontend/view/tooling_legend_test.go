package view_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/tooling"
)

// TestToolingLegend covers one entry under a chart.
func TestToolingLegend(t *testing.T) {
	t.Parallel()

	entry := view.NewTooling(&tooling.Report{Project: commandScope()}).Scopes[0].Windows[1].Legend[0]
	assert.Equal(t, "go test", entry.Name)
	assert.Equal(t, "s0", entry.Color)
	assert.Equal(t, int64(1), entry.Runs)
	assert.Equal(t, 7000*time.Millisecond, entry.Total)

	assert.Equal(t, "", view.ToolingLegend{}.Name)
}

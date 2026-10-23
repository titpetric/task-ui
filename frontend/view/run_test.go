package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestRun(t *testing.T) {
	t.Parallel()

	run := view.Run{Run: model.Run{Folder: "infra", Job: "plan", UserName: "tit petric", Forced: 1}, Live: true, Attached: 2}
	assert.Equal(t, "TP", run.Initials())
	assert.Equal(t, "infra/plan", run.Key())
	assert.True(t, run.IsForced())
	assert.False(t, run.IsScheduled())
	assert.True(t, run.Live)
	assert.Equal(t, 2, run.Attached)

	run.Scheduled, run.Forced = 1, 0
	assert.True(t, run.IsScheduled())
	assert.False(t, run.IsForced())
}

// A run whose recording expired says so in place of its output.
func TestRun_ExpiredText(t *testing.T) {
	t.Parallel()
	assert.Contains(t, view.Run{Expired: true}.ExpiredText(), "expired under the Data retention settings")
}

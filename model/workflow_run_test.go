package model_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// A finished workflow run lasted from start to finish; one in progress
// so far; one not started, nothing.
func TestWorkflowRun_Elapsed(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, 10, 19, 7, 0, 0, 0, time.UTC)
	end := start.Add(95 * time.Second)
	assert.Equal(t, 95*time.Second, (&model.WorkflowRun{StartedAt: &start, FinishedAt: &end}).Elapsed())
	running := time.Now().Add(-time.Minute)
	assert.True(t, (&model.WorkflowRun{StartedAt: &running}).Elapsed() >= time.Minute)
	assert.Equal(t, time.Duration(0), (&model.WorkflowRun{}).Elapsed())
	var none *model.WorkflowRun
	assert.Equal(t, time.Duration(0), none.Elapsed())
}

package view_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// A workflow as a page shows it: the row, its steps drawn as a chain,
// and what the user asking may do with it.
func TestWorkflow(t *testing.T) {
	t.Parallel()
	w := view.Workflow{
		Workflow: model.Workflow{ID: "w1", Name: "release"},
		Steps:    []view.Step{{Folder: "dockerfiles", Job: "build"}, {Folder: "dockerfiles", Job: "push"}},
		CanRun:   true,
	}
	assert.Equal(t, "release", w.Name)
	assert.Equal(t, "push", view.Chain(w.Steps)[1].Text)
	assert.True(t, w.CanRun)
	assert.False(t, w.CanChange)
}

// A run of a workflow carries the states of its steps.
func TestWorkflowRun(t *testing.T) {
	t.Parallel()
	r := view.WorkflowRun{
		WorkflowRun: model.WorkflowRun{ID: "wr1", State: model.StateFailed},
		Steps:       []view.Step{{Folder: "infra", Job: "plan", State: "ok"}, {Folder: "infra", Job: "apply", State: "failed"}},
	}
	assert.Equal(t, "wr1", r.ID)
	assert.Equal(t, "failed", view.Chain(r.Steps)[1].State)
}

func TestWorkflowRun_Took(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, 10, 20, 9, 0, 0, 0, time.UTC)
	end := start.Add(90 * time.Second)
	run := view.WorkflowRun{WorkflowRun: model.WorkflowRun{StartedAt: &start, FinishedAt: &end}}
	assert.Equal(t, 90*time.Second, run.Took())
}

package frontend

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// A workflow's steps carry their kind and the fields that kind draws:
// the workflow's name by its ref, and a check's markers and tries.
func TestViewSteps(t *testing.T) {
	t.Parallel()
	names := map[string]string{"w2": "deploy"}
	steps := viewSteps([]model.WorkflowStep{
		{Folder: "infra", Job: "plan"},
		{Kind: model.StepWorkflow, WorkflowRef: "w2"},
		{Kind: model.StepCheck, Markers: "Next:", Tries: 5},
	}, names)
	assert.Equal(t, model.StepJob, steps[0].Kind)
	assert.Equal(t, "deploy", steps[1].WorkflowName)
	assert.Equal(t, "Next:", steps[2].Markers)
	assert.Equal(t, 5, steps[2].Passes())
}

// A run's steps carry how each ended, the run it made, and what a check
// looked for.
func TestViewRunSteps(t *testing.T) {
	t.Parallel()
	rows := viewRunSteps([]model.WorkflowRunStep{
		{Folder: "infra", Job: "plan", State: model.StateOK, RunID: "r1"},
		{Kind: model.StepCheck, Markers: "Next:", State: model.StateOK},
	}, nil)
	assert.Equal(t, "r1", rows[0].RunID)
	assert.Equal(t, "Next:", rows[1].Markers)
	assert.Equal(t, "rotate", rows[1].Icon())
}

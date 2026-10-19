package model_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// A step runs a workflow in place of a job when it names one.
func TestWorkflowStep_IsWorkflow(t *testing.T) {
	t.Parallel()
	assert.True(t, model.WorkflowStep{WorkflowRef: "w1"}.IsWorkflow())
	assert.False(t, model.WorkflowStep{Folder: "infra", Job: "plan"}.IsWorkflow())
}

func TestWorkflowRunStep_IsWorkflow(t *testing.T) {
	t.Parallel()
	assert.True(t, model.WorkflowRunStep{WorkflowRef: "w1"}.IsWorkflow())
	assert.False(t, model.WorkflowRunStep{Folder: "infra", Job: "plan"}.IsWorkflow())
}

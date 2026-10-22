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

// A workflow runs with an enabled step or a prompt; its prompt runs in
// the root without a folder.
func TestWorkflow_Runs(t *testing.T) {
	t.Parallel()
	steps := []model.WorkflowStep{{Folder: "infra", Job: "plan", Disabled: true}}
	assert.False(t, (&model.Workflow{}).Runs(steps))
	assert.False(t, (&model.Workflow{}).Runs(nil))
	assert.False(t, (&model.Workflow{Prompt: "  "}).Runs(nil))
	assert.True(t, (&model.Workflow{Prompt: "say ok"}).Runs(steps))
	assert.True(t, (&model.Workflow{Prompt: "say ok"}).HasPrompt())
	assert.True(t, (&model.Workflow{}).Runs(append(steps, model.WorkflowStep{Folder: "infra", Job: "apply"})))
	assert.Equal(t, "_", (&model.Workflow{}).PromptPlace())
	assert.Equal(t, "malta/dockerfiles", (&model.Workflow{PromptFolder: "malta/dockerfiles"}).PromptPlace())
}

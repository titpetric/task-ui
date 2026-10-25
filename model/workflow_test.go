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

// A step's kind is its kind column, or for a row without one what its
// fields say: a prompt, a workflow, else a job. A prompt and a summary
// ask Claude in their folder and directory, the root without a folder.
func TestWorkflowStep_StepKind(t *testing.T) {
	t.Parallel()
	assert.Equal(t, model.StepJob, model.WorkflowStep{Folder: "infra", Job: "plan"}.StepKind())
	assert.Equal(t, model.StepWorkflow, model.WorkflowStep{WorkflowRef: "w1"}.StepKind())
	assert.Equal(t, model.StepPrompt, model.WorkflowStep{Prompt: "say ok"}.StepKind())
	assert.Equal(t, model.StepSummary, model.WorkflowStep{Kind: model.StepSummary, Folder: "_", Job: model.PromptJob}.StepKind())
	assert.Equal(t, model.StepSummary, model.WorkflowRunStep{Kind: model.StepSummary}.StepKind())
	assert.Equal(t, model.StepPrompt, model.WorkflowRunStep{Prompt: "say ok"}.StepKind())
}

func TestWorkflowStep_IsPrompt(t *testing.T) {
	t.Parallel()
	assert.True(t, model.WorkflowStep{Folder: "_", Job: model.PromptJob, Prompt: "say ok"}.IsPrompt())
	assert.True(t, model.WorkflowStep{Kind: model.StepSummary, Folder: "_", Job: model.PromptJob}.IsPrompt())
	assert.False(t, model.WorkflowStep{Folder: "infra", Job: "plan"}.IsPrompt())
	assert.True(t, model.WorkflowRunStep{Prompt: "say ok"}.IsPrompt())
	assert.True(t, model.WorkflowRunStep{Kind: model.StepSummary}.IsPrompt())
	assert.False(t, model.WorkflowRunStep{Folder: "infra", Job: "plan"}.IsPrompt())
	assert.Equal(t, "_", model.WorkflowStep{Prompt: "say ok"}.Place())
	assert.Equal(t, "malta/dockerfiles", model.WorkflowStep{Folder: "malta", Dir: "dockerfiles", Prompt: "say ok"}.Place())
}

func TestWorkflowStep_IsSummary(t *testing.T) {
	t.Parallel()
	assert.True(t, model.WorkflowStep{Kind: model.StepSummary}.IsSummary())
	assert.False(t, model.WorkflowStep{Prompt: "say ok"}.IsSummary())
	assert.True(t, model.WorkflowRunStep{Kind: model.StepSummary}.IsSummary())
	assert.False(t, model.WorkflowRunStep{Kind: model.StepPrompt, Prompt: "say ok"}.IsSummary())
}

// A workflow runs with an enabled step, and takes input with an enabled
// prompt or summary.
func TestWorkflow_Runs(t *testing.T) {
	t.Parallel()
	steps := []model.WorkflowStep{{Folder: "infra", Job: "plan", Disabled: true}}
	assert.False(t, (&model.Workflow{}).Runs(steps))
	assert.False(t, (&model.Workflow{}).Runs(nil))
	assert.True(t, (&model.Workflow{}).Runs(append(steps, model.WorkflowStep{Folder: "infra", Job: "apply"})))
}

func TestWorkflow_Asks(t *testing.T) {
	t.Parallel()
	steps := []model.WorkflowStep{{Folder: "infra", Job: "plan"}, {Kind: model.StepSummary, Folder: "_", Job: model.PromptJob, Disabled: true}}
	assert.False(t, (&model.Workflow{}).Asks(steps))
	steps[1].Disabled = false
	assert.True(t, (&model.Workflow{}).Asks(steps))
}

// A workflow runs as the user set on it, its owner without one.
func TestWorkflow_RunsAs(t *testing.T) {
	t.Parallel()
	wf := model.Workflow{UserID: "u-owner"}
	assert.Equal(t, "u-owner", wf.RunsAs())
	wf.RunAsUserID = "u-miguel"
	assert.Equal(t, "u-miguel", wf.RunsAs())
}

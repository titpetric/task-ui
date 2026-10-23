package view

import (
	"time"

	"github.com/titpetric/task-ui/model"
)

// Workflow is a workflow as a page shows it.
type Workflow struct {
	model.Workflow
	Steps []Step
	// Running is the ID of its run in progress, if any.
	Running string
	// LastRun is its newest run.
	LastRun *model.WorkflowRun
	// CanRun says whether the user may run every step.
	CanRun bool
	// CanChange says whether the user may change or delete it.
	CanChange bool
}

// Prompted reports whether a step of the workflow asks Claude, a prompt or a summary, so its Run takes input.
func (w Workflow) Prompted() bool {
	for _, step := range w.Steps {
		if kind := step.StepKind(); !step.Disabled && (kind == model.StepPrompt || kind == model.StepSummary) {
			return true
		}
	}
	return false
}

// PromptPlace is where the workflow's first prompt or summary step asks
// Claude, the place its thread shows before it starts: _ without one.
func (w Workflow) PromptPlace() string {
	for _, step := range w.Steps {
		if kind := step.StepKind(); kind == model.StepPrompt || kind == model.StepSummary {
			return model.WorkflowStep{Folder: step.Folder, Dir: step.Dir}.Place()
		}
	}
	return model.Default
}

// WorkflowRun is a run of a workflow with its steps.
type WorkflowRun struct {
	model.WorkflowRun
	Steps []Step
}

// Took is how long the run took so far, on a copy, which the pointer method Elapsed does not allow.
func (r WorkflowRun) Took() time.Duration {
	return r.WorkflowRun.Elapsed()
}

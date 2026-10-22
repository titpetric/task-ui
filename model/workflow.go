package model

import (
	"strings"
)

// The states a workflow run's step has beyond those of a run: it waits
// for the steps before it, and it skips when one of them did not
// end ok. The values mirror the CHECK in schema/workflow_run_step.up.sql.
const (
	// StepPending is a step whose turn has not come.
	StepPending State = "pending"
	// StepSkipped is a step that did not run: a disabled one, one after
	// a failure, or one after something ended the workflow run.
	StepSkipped State = "skipped"
)

// IsWorkflow says the step runs a workflow in place of a job.
func (s WorkflowStep) IsWorkflow() bool { return s.WorkflowRef != "" }

// IsWorkflow says the step ran a workflow in place of a job.
func (s WorkflowRunStep) IsWorkflow() bool { return s.WorkflowRef != "" }

// Runs reports whether the workflow has something to run with these
// steps: an enabled step, or a prompt.
func (w *Workflow) Runs(steps []WorkflowStep) bool {
	if w.HasPrompt() {
		return true
	}
	for _, step := range steps {
		if !step.Disabled {
			return true
		}
	}
	return false
}

// HasPrompt reports whether the workflow asks Claude once its steps end.
func (w *Workflow) HasPrompt() bool {
	return strings.TrimSpace(w.Prompt) != ""
}

// PromptPlace is where the workflow's prompt runs: PromptFolder, a
// folder or folder/dir, or the root, _, without one.
func (w *Workflow) PromptPlace() string {
	if w.PromptFolder == "" {
		return Default
	}
	return w.PromptFolder
}

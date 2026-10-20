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

// WorkflowRun is a run of a workflow with its steps.
type WorkflowRun struct {
	model.WorkflowRun
	Steps []Step
}

// Took is how long the run took, or has taken so far; a template calls it
// on a copy, which Elapsed, on a pointer, does not allow.
func (r WorkflowRun) Took() time.Duration {
	return r.WorkflowRun.Elapsed()
}

package view

import "github.com/titpetric/task-ui/model"

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

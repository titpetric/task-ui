package model

// The states a workflow run's step has beyond those of a run: it waits
// for the steps before it, and it is skipped when one of them did not
// end ok. The values mirror the CHECK in schema/workflow_run_step.up.sql.
const (
	// StepPending is a step whose turn has not come.
	StepPending State = "pending"
	// StepSkipped is a step that did not run, because one before it
	// failed or the workflow run was ended.
	StepSkipped State = "skipped"
)

// IsWorkflow says the step runs a workflow in place of a job.
func (s WorkflowStep) IsWorkflow() bool { return s.WorkflowRef != "" }

// IsWorkflow says the step ran a workflow in place of a job.
func (s WorkflowRunStep) IsWorkflow() bool { return s.WorkflowRef != "" }

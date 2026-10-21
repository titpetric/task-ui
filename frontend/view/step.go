package view

// Step is a step of a workflow, or of a run of one, as a page shows
// it: a job of a folder, or another workflow.
type Step struct {
	Folder string
	Job    string
	// WorkflowID and WorkflowName are the workflow a step runs in place
	// of a job.
	WorkflowID   string
	WorkflowName string
	// State is how a run's step ended; RunID the run it made of a job,
	// WorkflowRunID the run it made of a workflow.
	State         string
	RunID         string
	WorkflowRunID string
}

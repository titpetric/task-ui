package api

// StepRequest names one step of a workflow: a job of a folder, or a
// workflow to run in its place.
type StepRequest struct {
	Folder string `json:"folder,omitempty"`
	Job    string `json:"job,omitempty"`
	// Workflow is the ID of a workflow the step runs instead of a job.
	Workflow string `json:"workflow,omitempty"`
}

package api

// StepRequest names one step of a workflow: a job of a folder, or a
// workflow to run in its place.
type StepRequest struct {
	Folder string `json:"folder,omitempty"`
	Job    string `json:"job,omitempty"`
	// Workflow is the ID of a workflow the step runs; a step runs a job or
	// a workflow.
	Workflow string `json:"workflow,omitempty"`
	// Enabled runs the step, the default; false keeps it in the workflow
	// and a run skips it.
	Enabled *bool `json:"enabled,omitempty"`
	// Args follow the job's argv after --, for a job that takes args:
	// the command of @host:<host>, which needs one.
	Args RunArgs `json:"args,omitempty"`
}

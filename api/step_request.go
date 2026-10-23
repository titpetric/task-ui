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
	// Target and Args are what a step of @host runs: a host of the
	// inventory, or * for every host with an ssh login, and the command.
	Target string  `json:"target,omitempty"`
	Args   RunArgs `json:"args,omitempty"`
}

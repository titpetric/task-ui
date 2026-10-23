package api

import (
	"github.com/titpetric/task-ui/model"
)

// StepRequest names one step of a workflow by its kind: a job of a
// folder, a workflow to run in its place, a prompt to ask Claude, or a
// summary, Claude's summary of the steps before it.
type StepRequest struct {
	// Kind is job, workflow, prompt or summary; left out, the fields
	// given say: prompt with a prompt, workflow with a workflow, else
	// job. A summary step says so.
	Kind   string `json:"kind,omitempty"`
	Folder string `json:"folder,omitempty"`
	Job    string `json:"job,omitempty"`
	// Workflow is the ID of a workflow the step runs.
	Workflow string `json:"workflow,omitempty"`
	// Enabled runs the step, the default; false keeps it in the workflow
	// and a run skips it.
	Enabled *bool `json:"enabled,omitempty"`
	// Args follow the job's argv after --, for a job that takes args:
	// the command of @host:<host>, which needs one.
	Args RunArgs `json:"args,omitempty"`
	// Prompt is what a prompt step asks Claude, in Folder, _ or empty
	// for the root, and Dir inside it, with the output of the steps
	// before it; a summary step asks in the same place, with no prompt.
	// Both take prompt/edit.
	Prompt string `json:"prompt,omitempty"`
	Dir    string `json:"dir,omitempty"`
}

// StepKind is the kind the request names, or the one its fields say.
func (s StepRequest) StepKind() string {
	return model.WorkflowStep{Kind: s.Kind, Prompt: s.Prompt, WorkflowRef: s.Workflow}.StepKind()
}

package workflow

// StepEvent is the payload of a workflow's `step` event: the step as
// its row stands.
type StepEvent struct {
	RunID       string `json:"run_id"`
	Position    int64  `json:"position"`
	Folder      string `json:"folder,omitempty"`
	Job         string `json:"job,omitempty"`
	WorkflowRef string `json:"workflow_ref,omitempty"`
	// Kind is the step's kind; Prompt says it asks Claude in Folder, a
	// prompt or a summary step.
	Kind           string `json:"kind"`
	Prompt         bool   `json:"prompt,omitempty"`
	State          string `json:"state"`
	StepRunID      string `json:"step_run_id,omitempty"`
	WorkflowRunRef string `json:"workflow_run_ref,omitempty"`
}

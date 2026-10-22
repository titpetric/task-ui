package api

// WorkflowRunRequest starts a workflow run: Input goes to the
// workflow's prompt, after it under "Input:". The body is optional.
type WorkflowRunRequest struct {
	Input string `json:"input,omitempty"`
}

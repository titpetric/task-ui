package api

// PromptRequest runs a prompt against a workspace: Claude Code reads it
// in the workspace's checkout.
type PromptRequest struct {
	Prompt string `json:"prompt"`
	Cols   int    `json:"cols"`
	Rows   int    `json:"rows"`
}

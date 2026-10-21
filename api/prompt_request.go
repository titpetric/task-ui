package api

// PromptRequest runs a prompt: Claude Code reads it in a folder, or in
// Dir inside it, for POST /api/prompts, or in a workspace's checkout,
// which its route names.
type PromptRequest struct {
	Folder string `json:"folder"`
	Dir    string `json:"dir"`
	Prompt string `json:"prompt"`
	Cols   int    `json:"cols"`
	Rows   int    `json:"rows"`
}

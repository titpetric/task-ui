package api

// PromptRequest puts a prompt to Claude: Claude Code reads it in a
// folder, or in Dir inside it, for POST /api/prompts, or in a workspace's
// checkout, which its route names. Parent makes it a reply in that
// prompt's thread, which asks again in the thread's place.
type PromptRequest struct {
	Folder string `json:"folder"`
	Dir    string `json:"dir"`
	Prompt string `json:"prompt"`
	Parent string `json:"parent"`
	Cols   int    `json:"cols"`
	Rows   int    `json:"rows"`
}

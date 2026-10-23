package api

// PromptRequest puts a prompt to Claude: Claude Code reads it in a
// folder, or in Dir inside it, for POST /api/prompts, or in a workspace's
// checkout, which its route names. Parent makes it a reply in that
// prompt's thread, which asks again in the thread's place. Run asks
// about a run instead, in its folder and directory, with its output.
type PromptRequest struct {
	Folder string `json:"folder"`
	Dir    string `json:"dir"`
	Run    string `json:"run"`
	Prompt string `json:"prompt"`
	Parent string `json:"parent"`
	// Area asks from an area's page in place of a folder's: workflows,
	// the Workflows page's box, which builds and changes workflows with
	// the task tools, in Folder and Dir, the root unless given.
	Area string `json:"area"`
	Cols int    `json:"cols"`
	Rows int    `json:"rows"`
}

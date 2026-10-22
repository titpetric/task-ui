package view

// Prompt is the prompt box of a folder page or of a directory browsed
// inside a folder: Claude Code on a prompt, run as internal/claude-prompt
// in that directory, and the threads of the prompts asked there, each
// question with Claude's answer.
type Prompt struct {
	// API is the route the box posts to: /api/prompts, or the workspace
	// route of the default workspace, which is no folder.
	API string
	// Folder and Dir name the directory for /api/prompts.
	Folder string
	Dir    string
	// Where is the directory as the box names it, ending in /.
	Where string
	// Open shows the box open, as on a workspace's page.
	Open bool
	// Place is where the prompts run as their notes hold it: the folder
	// and the directory, or _ for the root.
	Place string
	// Threads are the prompts asked in Place, newest first.
	Threads []Thread
}

// PromptIn returns the prompt box of a directory inside a folder, the
// folder itself for an empty dir.
func PromptIn(folderID, dir string) *Prompt {
	place := folderID
	if dir != "" {
		place += "/" + dir
	}
	return &Prompt{API: "/api/prompts", Folder: folderID, Dir: dir, Where: place + "/", Place: place}
}

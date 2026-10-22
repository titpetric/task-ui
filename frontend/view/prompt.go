package view

// Prompt is the prompt box of a folder page, of a directory browsed
// inside a folder, or of a run's page: Claude Code on a prompt, a run
// of @claude-prompt in that directory, and the threads of the prompts asked there, each
// question with Claude's answer.
type Prompt struct {
	// API is the route the box posts to: /api/prompts, or the workspace
	// route of the default workspace, which is no folder.
	API string
	// Folder and Dir name the directory for /api/prompts.
	Folder string
	Dir    string
	// Run is the run a run page's box asks about, with its output, in
	// the run's directory; its threads are those about the run.
	Run string
	// WorkflowRun is the workflow run whose prompt's thread the box of a
	// workflow run's page lists, without a box to ask.
	WorkflowRun string
	// Watch lists the threads without the box to ask, for a user who
	// reads them only.
	Watch bool
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

// PromptOfWorkflowRun returns the box of a workflow run's page: the
// thread of the workflow's prompt, asked in place once the steps
// ended, where names place.
func PromptOfWorkflowRun(runID, place, where string) *Prompt {
	return &Prompt{WorkflowRun: runID, Watch: true, Open: true, Where: where, Place: place}
}

// PromptAbout returns the prompt box of a run's page, about the run in
// where, its folder and directory.
func PromptAbout(runID, where string) *Prompt {
	return &Prompt{API: "/api/prompts", Run: runID, Where: where + "/", Place: where, Open: true}
}

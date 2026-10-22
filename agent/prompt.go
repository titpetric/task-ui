package agent

import (
	"context"
	"errors"
	"io"
	"os"
)

// Prompt runs Claude Code on a prompt, TASK_UI_PROMPT, in the directory
// TASK_UI_PROMPT_DIR names, a folder or a directory inside one, or else
// in the workspace's checkout, TASK_UI_WORKSPACE_DIR: Claude reads the
// CLAUDE.md and .claude/ there and above. The prompt is a thread, oldest
// first, each message under its author in a thread of more than one; the
// last is the question. Claude's answer goes to the terminal and to the file
// TASK_UI_PROMPT_REPLY names, which the board replies with.
func (a *Agent) Prompt(ctx context.Context) error {
	thread := a.getenv("TASK_UI_PROMPT")
	if thread == "" {
		return errors.New("TASK_UI_PROMPT: this job runs from a prompt against a workspace")
	}
	dir := a.getenv("TASK_UI_PROMPT_DIR")
	if dir == "" {
		dir = a.getenv("TASK_UI_WORKSPACE_DIR")
	}
	if dir == "" {
		return errors.New("TASK_UI_PROMPT_DIR or TASK_UI_WORKSPACE_DIR: this job runs in a folder or a workspace")
	}
	if a.getenv("CLAUDE_CODE_OAUTH_TOKEN") == "" {
		return errNoToken
	}
	prompt := "A prompt from a task-ui folder page. The thread, oldest first; with several messages each is under its author, your earlier answers among them, and the last message is the question:\n\n" +
		thread + "\n\n" +
		"Your final message is the answer, in markdown, shown under the question on the page."

	out := a.Stdout
	if path := a.getenv("TASK_UI_PROMPT_REPLY"); path != "" {
		reply, err := os.Create(path)
		if err != nil {
			return err
		}
		defer reply.Close()
		out = io.MultiWriter(a.Stdout, reply)
	}
	return a.ask(ctx, dir, prompt, out)
}

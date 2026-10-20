package agent

import (
	"context"
	"errors"
)

// Prompt runs Claude Code on a prompt against a workspace, TASK_UI_PROMPT,
// in the workspace's checkout, TASK_UI_WORKSPACE_DIR: Claude reads the
// workspace's CLAUDE.md and .claude/ there, and writes its reply to the
// terminal.
func (a *Agent) Prompt(ctx context.Context) error {
	prompt := a.getenv("TASK_UI_PROMPT")
	if prompt == "" {
		return errors.New("TASK_UI_PROMPT: this job runs from a prompt against a workspace")
	}
	dir := a.getenv("TASK_UI_WORKSPACE_DIR")
	if dir == "" {
		return errors.New("TASK_UI_WORKSPACE_DIR: this job runs in a workspace")
	}
	return a.ask(ctx, dir, prompt, a.Stdout)
}

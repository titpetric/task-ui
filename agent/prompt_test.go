package agent_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/titpetric/task-ui/agent"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestAgent_Prompt(t *testing.T) {
	t.Parallel()
	workspace := realDir(t)
	var out bytes.Buffer
	a := &agent.Agent{Stdout: &out, Stderr: io.Discard, Env: []string{
		"PATH=" + fakeClaude(t), "HOME=/home/server", "CLAUDE_CODE_OAUTH_TOKEN=sk-1",
		"TASK_UI_PROMPT=List the skills.", "TASK_UI_WORKSPACE_DIR=" + workspace,
	}}
	assert.NoError(t, a.Run(context.Background(), []string{"prompt"}))
	text := out.String()
	assert.Contains(t, text, "cwd="+workspace+"\n")
	assert.Contains(t, text, "home=/home/server\n", "the server's HOME stays")
	assert.Contains(t, text, "arg=--output-format\narg=text\narg=List the skills.\n")

	// TASK_UI_PROMPT_DIR, a folder or a directory inside one, wins over
	// the workspace.
	here := realDir(t)
	out.Reset()
	a.Env = append(a.Env, "TASK_UI_PROMPT_DIR="+here)
	assert.NoError(t, a.Run(context.Background(), []string{"prompt"}))
	assert.Contains(t, out.String(), "cwd="+here+"\n")

	assert.ErrorContains(t, (&agent.Agent{Env: []string{"TASK_UI_WORKSPACE_DIR=" + workspace}}).Prompt(context.Background()), "TASK_UI_PROMPT")
	assert.ErrorContains(t, (&agent.Agent{Env: []string{"TASK_UI_PROMPT=x"}}).Prompt(context.Background()), "TASK_UI_PROMPT_DIR or TASK_UI_WORKSPACE_DIR")
	assert.ErrorContains(t, (&agent.Agent{Env: []string{"TASK_UI_PROMPT=x", "TASK_UI_WORKSPACE_DIR=" + workspace}}).Prompt(context.Background()), "CLAUDE_CODE_OAUTH_TOKEN")
}

package agent_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
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
	assert.Contains(t, text, "arg=--output-format\narg=text\narg=A prompt from a task-ui folder page.")
	assert.Contains(t, text, "the last message is the question:\n\nList the skills.\n\nYour final message is the answer, in markdown")

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

// The thread reaches Claude whole, and its answer goes to the reply file
// as well as the terminal.
func TestAgent_Prompt_reply(t *testing.T) {
	t.Parallel()
	here := realDir(t)
	reply := filepath.Join(t.TempDir(), "reply.md")
	var out bytes.Buffer
	a := &agent.Agent{Stdout: &out, Stderr: io.Discard, Env: []string{
		"PATH=" + fakeClaude(t), "CLAUDE_CODE_OAUTH_TOKEN=sk-1", "TASK_UI_PROMPT_DIR=" + here,
		"TASK_UI_PROMPT=admin:\nwhat builds here\n\nclaude:\nthe images\n\nadmin:\nwhich first", "TASK_UI_PROMPT_REPLY=" + reply,
	}}
	assert.NoError(t, a.Run(context.Background(), []string{"prompt"}))
	assert.Contains(t, out.String(), "admin:\nwhat builds here\n\nclaude:\nthe images\n\nadmin:\nwhich first\n\nYour final message")
	written, err := os.ReadFile(reply)
	assert.NoError(t, err)
	assert.Equal(t, out.String(), string(written))
}

// With TASK_UI_HOSTS, Claude's system prompt gets the hosts and how to
// reach them over ssh; without it, there is no system prompt.
func TestAgent_Prompt_hosts(t *testing.T) {
	t.Parallel()
	here := realDir(t)
	var out bytes.Buffer
	env := []string{"PATH=" + fakeClaude(t), "CLAUDE_CODE_OAUTH_TOKEN=sk-1", "TASK_UI_PROMPT_DIR=" + here, "TASK_UI_PROMPT=list the containers"}
	a := &agent.Agent{Stdout: &out, Stderr: io.Discard, Env: append(env, "TASK_UI_HOSTS=web\tweb.example\troot\tup\ndb\tdb\t\tdown")}
	assert.NoError(t, a.Run(context.Background(), []string{"prompt"}))
	text := out.String()
	assert.Contains(t, text, "arg=--append-system-prompt\narg=task-ui's inventory lists these hosts")
	assert.Contains(t, text, "\n\nweb\tweb.example\troot\tup\ndb\tdb\t\tdown\n\n")
	assert.Contains(t, text, "ssh -o BatchMode=yes -o ConnectTimeout=10 -o StrictHostKeyChecking=accept-new <user>@<host> '<command>'")
	assert.Contains(t, text, "run them in parallel")

	out.Reset()
	a.Env = append(env, "TASK_UI_HOSTS=The inventory could not be read: hub down")
	assert.NoError(t, a.Run(context.Background(), []string{"prompt"}))
	assert.Contains(t, out.String(), "arg=--append-system-prompt\narg=task-ui's inventory of hosts: The inventory could not be read: hub down\n")

	out.Reset()
	a.Env = env
	assert.NoError(t, a.Run(context.Background(), []string{"prompt"}))
	assert.NotContains(t, out.String(), "--append-system-prompt")
}

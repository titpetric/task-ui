package agent_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/agent"
	"github.com/titpetric/task-ui/claude"
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

// A thread from a run's page reaches Claude after the run, from the
// file TASK_UI_PROMPT_RUN names, with how to treat a fix; a file that is
// not there fails the prompt.
func TestAgent_Prompt_run(t *testing.T) {
	t.Parallel()
	here := realDir(t)
	run := filepath.Join(t.TempDir(), "run.md")
	assert.NoError(t, os.WriteFile(run, []byte("Run 01 of the job build in infra, by ops.\nState: failed, exit code 2.\n"), 0o644))
	var out bytes.Buffer
	env := []string{"PATH=" + fakeClaude(t), "CLAUDE_CODE_OAUTH_TOKEN=sk-1", "TASK_UI_PROMPT_DIR=" + here, "TASK_UI_PROMPT=why did it fail"}
	a := &agent.Agent{Stdout: &out, Stderr: io.Discard, Env: append(env, "TASK_UI_PROMPT_RUN="+run)}
	assert.NoError(t, a.Run(context.Background(), []string{"prompt"}))
	text := out.String()
	assert.Contains(t, text, "arg=A prompt from the page of a task-ui run, about that run. The run:\n\nRun 01 of the job build in infra, by ops.\nState: failed, exit code 2.\n\nThe question is about this run.")
	assert.Contains(t, text, "commit it to main and push it")
	assert.Contains(t, text, "the last message is the question:\n\nwhy did it fail\n\nYour final message is the answer, in markdown, shown under the question on the run's page.")

	a.Env = append(env, "TASK_UI_PROMPT_RUN="+filepath.Join(here, "gone.md"))
	assert.Error(t, a.Run(context.Background(), []string{"prompt"}))
}

// TASK_UI_PROMPT_INSTRUCTIONS goes to the system prompt, ahead of the
// hosts, and the question stays the thread's.
func TestAgent_Prompt_instructions(t *testing.T) {
	t.Parallel()
	here := realDir(t)
	var out bytes.Buffer
	a := &agent.Agent{Stdout: &out, Stderr: io.Discard, Env: []string{
		"PATH=" + fakeClaude(t), "CLAUDE_CODE_OAUTH_TOKEN=sk-1", "TASK_UI_PROMPT_DIR=" + here, "TASK_UI_PROMPT=Summarize this run.",
		"TASK_UI_PROMPT_INSTRUCTIONS=" + claude.RunSummary, "TASK_UI_HOSTS=web",
	}}
	assert.NoError(t, a.Run(context.Background(), []string{"prompt"}))
	text := out.String()
	assert.Contains(t, text, "arg=--append-system-prompt\narg="+claude.RunSummary+"\n\ntask-ui's inventory of hosts: web\n")
	assert.Contains(t, text, "the last message is the question:\n\nSummarize this run.\n\n")
}

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
	assert.Contains(t, text, "arg=--output-format\narg=text\narg=--append-system-prompt\narg=# task-ui\n")
	assert.Contains(t, text, "\narg=A prompt from a task-ui folder page.")
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
// reach them over ssh; without it, the system prompt is the context.
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
	assert.Contains(t, out.String(), "arg=--append-system-prompt\narg=# task-ui\n")
	assert.NotContains(t, out.String(), "task-ui's inventory")
}

// The system prompt ends with the context: the run's artifact
// directory, the user with the role and what the user may do, and
// context.md.
func TestAgent_Prompt_context(t *testing.T) {
	t.Parallel()
	here := realDir(t)
	var out bytes.Buffer
	a := &agent.Agent{Stdout: &out, Stderr: io.Discard, Env: []string{
		"PATH=" + fakeClaude(t), "CLAUDE_CODE_OAUTH_TOKEN=sk-1", "TASK_UI_PROMPT_DIR=" + here, "TASK_UI_PROMPT=what are artifacts",
		"TASK_UI_ARTIFACTS=/data/artifacts/01", "TASK_UI_USER=claude", "TASK_UI_ROLE=runner", "TASK_UI_PERMISSIONS=note/view, run; everything else is denied",
	}}
	assert.NoError(t, a.Run(context.Background(), []string{"prompt"}))
	text := out.String()
	assert.Contains(t, text, "arg=--append-system-prompt\narg=This run's artifact directory is $TASK_UI_ARTIFACTS, /data/artifacts/01:")
	assert.Contains(t, text, "\n\n"+claude.Context("claude", "runner", "note/view, run; everything else is denied")+"\n")
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

// A workflow run's prompt reaches Claude after the run's steps, from
// the file TASK_UI_PROMPT_WORKFLOW names, told to change nothing the
// prompt does not name; a file that is not there fails the prompt.
func TestAgent_Prompt_workflow(t *testing.T) {
	t.Parallel()
	here := realDir(t)
	run := filepath.Join(t.TempDir(), "workflow.md")
	assert.NoError(t, os.WriteFile(run, []byte("Workflow run 01 of the workflow ship, by ops, with 1 step.\n"), 0o644))
	var out bytes.Buffer
	env := []string{"PATH=" + fakeClaude(t), "CLAUDE_CODE_OAUTH_TOKEN=sk-1", "TASK_UI_PROMPT_DIR=" + here, "TASK_UI_PROMPT=summarize, change nothing"}
	a := &agent.Agent{Stdout: &out, Stderr: io.Discard, Env: append(env, "TASK_UI_PROMPT_WORKFLOW="+run)}
	assert.NoError(t, a.Run(context.Background(), []string{"prompt"}))
	text := out.String()
	assert.Contains(t, text, "arg=A prompt of a task-ui workflow, asked in a run of the workflow once the steps before it ended, about that run. The run:\n\nWorkflow run 01 of the workflow ship, by ops, with 1 step.\n\nChange nothing unless the prompt asks for a change")
	assert.Contains(t, text, "the first message is the workflow's prompt:\n\nsummarize, change nothing\n\nYour final message is the answer")

	a.Env = append(env, "TASK_UI_PROMPT_WORKFLOW="+filepath.Join(here, "gone.md"))
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

// A run with an MCP configuration passes its file to Claude
// and tells it task-ui's tools are there as the run's user; a run
// without has neither.
func TestAgent_Prompt_mcp(t *testing.T) {
	t.Parallel()
	here := realDir(t)
	var out bytes.Buffer
	env := []string{"PATH=" + fakeClaude(t), "CLAUDE_CODE_OAUTH_TOKEN=sk-1", "TASK_UI_PROMPT_DIR=" + here, "TASK_UI_PROMPT=build a workflow", "TASK_UI_USER=claude"}
	a := &agent.Agent{Stdout: &out, Stderr: io.Discard, Env: append(env, "TASK_UI_MCP_CONFIG=/run/mcp/mcp.json")}
	assert.NoError(t, a.Run(context.Background(), []string{"prompt"}))
	assert.Contains(t, out.String(), "arg=--mcp-config\narg=/run/mcp/mcp.json\n")
	assert.Contains(t, out.String(), "task-ui's MCP server, task, is connected for this run as claude, with that user's permissions and no more")

	out.Reset()
	a.Env = env
	assert.NoError(t, a.Run(context.Background(), []string{"prompt"}))
	assert.NotContains(t, out.String(), "--mcp-config")
	assert.NotContains(t, out.String(), "is connected for this run")
}

// With TASK_UI_PROMPT_SOURCE, the system prompt names the thread to
// reply in over MCP, and the final message is the fallback reply.
func TestAgent_Prompt_source(t *testing.T) {
	t.Parallel()
	here := realDir(t)
	var out bytes.Buffer
	a := &agent.Agent{Stdout: &out, Stderr: io.Discard, Env: []string{
		"PATH=" + fakeClaude(t), "CLAUDE_CODE_OAUTH_TOKEN=sk-1", "TASK_UI_PROMPT_DIR=" + here, "TASK_UI_PROMPT=build it",
		"TASK_UI_PROMPT_SOURCE=## The prompt's source\n\nnote_comment with the id n1",
	}}
	assert.NoError(t, a.Run(context.Background(), []string{"prompt"}))
	text := out.String()
	assert.Contains(t, text, "arg=--append-system-prompt\narg=## The prompt's source\n\nnote_comment with the id n1\n\n")
	assert.Contains(t, text, "build it\n\nPost the answer as a reply in the thread, as the prompt's source in the system prompt says; your final message is the fallback")
}

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

func TestAgent_Note(t *testing.T) {
	t.Parallel()
	root := realDir(t)
	reply := filepath.Join(t.TempDir(), "reply.md")
	var out bytes.Buffer
	a := &agent.Agent{Stdout: &out, Stderr: io.Discard, Env: []string{
		"PATH=" + fakeClaude(t), "HOME=/home/server", "CLAUDE_CODE_OAUTH_TOKEN=sk-1", "TASK_UI_ROOT=" + root,
		"TASK_UI_NOTE=ops:\n@claude bump it", "TASK_UI_NOTE_ID=n1", "TASK_UI_NOTE_AUTHOR=ops", "TASK_UI_NOTE_REPLY=" + reply,
	}}
	assert.NoError(t, a.Run(context.Background(), []string{"note"}))
	text := out.String()
	assert.Contains(t, text, "cwd="+root+"\n")
	assert.Contains(t, text, "sandbox=1\n")
	assert.Contains(t, text, "home=/home/server\n")
	assert.Contains(t, text, "arg=-p\narg=--dangerously-skip-permissions\narg=--output-format\narg=text\narg=An instruction on the task-ui Notes board, note n1 by ops.")
	assert.Contains(t, text, "ops:\n@claude bump it\n\nThe working directory holds the projects")

	// The reply file holds what the terminal shows.
	written, err := os.ReadFile(reply)
	assert.NoError(t, err)
	assert.Equal(t, text, string(written))
}

func TestAgent_Note_refused(t *testing.T) {
	t.Parallel()
	path := fakeClaude(t)
	var out bytes.Buffer
	none := &agent.Agent{Stdout: &out, Env: []string{"PATH=" + path, "CLAUDE_CODE_OAUTH_TOKEN=sk-1"}}
	assert.ErrorContains(t, none.Note(context.Background()), "TASK_UI_NOTE: this job runs from a note on the Notes board")
	untokened := &agent.Agent{Stdout: &out, Env: []string{"PATH=" + path, "TASK_UI_NOTE=@claude hi"}}
	assert.ErrorContains(t, untokened.Note(context.Background()), "CLAUDE_CODE_OAUTH_TOKEN: set up Settings, Integrations, Claude Code, and name this user in Who")
	assert.Equal(t, "", out.String())

}

package agent_test

import (
	"testing"

	"github.com/titpetric/task-ui/agent"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestFolder(t *testing.T) {
	t.Parallel()
	f := agent.Folder("/srv/root", "claude")
	assert.Equal(t, agent.Namespace, f.ID)
	assert.Equal(t, "internal", f.ID)
	assert.Equal(t, "/srv/root", f.Dir)
	assert.Len(t, f.Jobs, 3)
	assert.Len(t, f.SourceJobs(f.Sources[0]), 3, "the folder page lists the jobs under the source")

	note := f.Job("claude-note")
	assert.Equal(t, "internal/claude-note", note.Key())
	assert.Equal(t, []string{"claude"}, note.Allow)
	assert.Equal(t, []string{"agent", "note"}, note.Command[1:])
	assert.Equal(t, []string{"admin"}, f.Job("claude-prompt").Allow)
	assert.Equal(t, []string{"agent", "prompt"}, f.Job("claude-prompt").Command[1:])
	shell := f.Job("claude")
	assert.True(t, shell.Interactive)
	assert.Equal(t, []string{"admin"}, shell.Allow)
	assert.Equal(t, []string{"agent", "shell"}, shell.Command[1:])
	assert.NotEmpty(t, shell.Description)

	// Without an agent, the notes job is the admins'.
	assert.Equal(t, []string{"admin"}, agent.Folder("/srv/root", "").Job("claude-note").Allow)
}

package agent

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
)

// The tooling of TASK_UI_TOOLING reads as the tools the run has and
// where to ask one for its usage; nothing without tools.
func TestToolset(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "", toolset(""))
	tools := toolset("git: version control\ngo: Go toolchain")
	assert.Contains(t, tools, "available to you here")
	assert.Contains(t, tools, "git: version control")
	assert.Contains(t, tools, "go: Go toolchain")
	assert.Contains(t, tools, "`<tool> --help`")
}

// The context of a run carries the tooling and its usage hint.
func TestAgent_ContextTooling(t *testing.T) {
	t.Parallel()
	a := &Agent{Env: []string{"TASK_UI_TOOLING=git: version control"}}
	context := a.context()
	assert.Contains(t, context, "git: version control")
	assert.Contains(t, context, "`<tool> --help`")

	empty := (&Agent{}).context()
	assert.NotContains(t, empty, "available to you here")
}

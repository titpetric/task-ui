package agent_test

import (
	"context"
	"errors"
	"io"
	"os/exec"
	"testing"

	"github.com/titpetric/task-ui/agent"
	"github.com/titpetric/task-ui/internal/assert"
)

// Claude Code comes from the PATH of the environment, and its exit code
// is the command's.
func TestAgent_claude(t *testing.T) {
	t.Parallel()
	missing := &agent.Agent{Stdout: io.Discard, Env: []string{"PATH=" + t.TempDir(), "CLAUDE_CODE_OAUTH_TOKEN=sk-1", "TASK_UI_NOTE=@claude hi"}}
	assert.ErrorContains(t, missing.Note(context.Background()), "claude: not found")

	failing := &agent.Agent{Stdout: io.Discard, Stderr: io.Discard, Env: []string{"PATH=" + fakeClaude(t), "CLAUDE_CODE_OAUTH_TOKEN=sk-1", "TASK_UI_NOTE=@claude hi", "FAKE_EXIT=3"}}
	var exit *exec.ExitError
	assert.True(t, errors.As(failing.Note(context.Background()), &exit))
	assert.Equal(t, 3, exit.ExitCode())
}

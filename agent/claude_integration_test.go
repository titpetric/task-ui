//go:build integration

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

	// OpenRouter's auth token and base URL sign Claude Code in too.
	openrouter := &agent.Agent{Stdout: io.Discard, Stderr: io.Discard, Env: []string{"PATH=" + fakeClaude(t), "ANTHROPIC_AUTH_TOKEN=sk-or-1", "ANTHROPIC_BASE_URL=https://openrouter.ai/api", "TASK_UI_NOTE=@claude hi"}}
	assert.NoError(t, openrouter.Note(context.Background()))
	unrouted := &agent.Agent{Stdout: io.Discard, Env: []string{"PATH=" + fakeClaude(t), "ANTHROPIC_AUTH_TOKEN=sk-or-1", "TASK_UI_NOTE=@claude hi"}}
	assert.ErrorContains(t, unrouted.Note(context.Background()), "CLAUDE_CODE_OAUTH_TOKEN or ANTHROPIC_AUTH_TOKEN")
}

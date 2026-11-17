//go:build integration

package agent

import (
	"context"
	"io"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
)

// Prompt refuses an empty thread, a missing directory and a run with no token.
func TestAgent_Prompt(t *testing.T) {
	t.Parallel()
	refused := func(env ...string) error {
		return (&Agent{Env: env, Stdout: io.Discard, Stderr: io.Discard}).Prompt(context.Background())
	}
	assert.Error(t, refused())
	assert.Error(t, refused("TASK_UI_PROMPT=what is it"))
	assert.Equal(t, errNoToken, refused("TASK_UI_PROMPT=what is it", "TASK_UI_PROMPT_DIR=/tmp"))
}

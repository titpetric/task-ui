package agent_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/agent"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestAgent_Run(t *testing.T) {
	t.Parallel()
	a := &agent.Agent{}
	assert.ErrorContains(t, a.Run(context.Background(), nil), "usage: task-ui agent")
	assert.ErrorContains(t, a.Run(context.Background(), []string{"nope"}), "usage: task-ui agent")
}

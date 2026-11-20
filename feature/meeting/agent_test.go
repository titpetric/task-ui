package meeting

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/assert"
)

// TestAgentHooks keeps the three callbacks a board supplies for a room's agent.
func TestAgentHooks(t *testing.T) {
	t.Parallel()
	hooks := &AgentHooks{
		Valid: func(ctx context.Context, name string) (*model.User, error) {
			return &model.User{Name: name}, nil
		},
		Ask: func(context.Context, *model.User, string, string, string, string) (string, error) {
			return "note", nil
		},
		Watch: func(string) (<-chan model.Note, func()) {
			return nil, func() {}
		},
	}
	agent, err := hooks.Valid(context.Background(), "claude")
	assert.NoError(t, err)
	assert.Equal(t, "claude", agent.Name)
	assert.True(t, hooks.Ask != nil && hooks.Watch != nil)
}

package runner_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/runner"
)

// Integrations is the contract between the manager and the integrations.
func TestIntegrations(t *testing.T) {
	t.Parallel()
	var integrations runner.Integrations = fixedIntegrations{user: "ops"}
	vars, err := integrations.Env(context.Background(), "ops", "runner")
	assert.NoError(t, err)
	assert.Equal(t, "sk-1", vars["CLAUDE_CODE_OAUTH_TOKEN"])
}

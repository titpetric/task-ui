package integration_test

import (
	"testing"

	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/internal/assert"
)

// Configured holds without the optional settings: Claude Code without
// Who, which then reaches every run.
func TestConfigured(t *testing.T) {
	t.Parallel()
	claude := *integration.KindOf(integration.Claude)
	assert.True(t, integration.Configured(claude, map[string]string{"token": "sk-1"}))
	assert.False(t, integration.Configured(claude, map[string]string{"who": "admin"}))
}

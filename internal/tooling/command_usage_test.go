package tooling_test

import (
	"encoding/json"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/tooling"
)

// TestCommandUsage reads the per-command series as atkins prints it.
func TestCommandUsage(t *testing.T) {
	t.Parallel()

	var scope tooling.Scope
	data := `{"command_usage": [{"date": "2026-10-23", "command": "docker compose up", "tool": "docker", "runs": 2, "total_ms": 7000}]}`
	assert.NoError(t, json.Unmarshal([]byte(data), &scope))

	assert.Len(t, scope.CommandUsage, 1)
	assert.Equal(t, tooling.CommandUsage{Date: "2026-10-23", Command: "docker compose up", Tool: "docker", Runs: 2, TotalMs: 7000}, scope.CommandUsage[0])
}

package tooling_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/tooling"
)

// TestSample reads a command as `atkins --stats --json --all` prints it.
func TestSample(t *testing.T) {
	t.Parallel()

	var scope tooling.Scope
	data := `{"samples": [{"at": "2026-10-23T16:23:40.338Z", "command": "docker compose build", "tool": "docker", "kind": "step", "duration_ms": 2789, "exit_code": 1, "failed": true, "pipeline": "Image", "job": "build", "invocation_id": "01M5"}]}`
	assert.NoError(t, json.Unmarshal([]byte(data), &scope))

	assert.Len(t, scope.Samples, 1)
	sample := scope.Samples[0]
	assert.True(t, sample.At.Equal(time.Date(2026, 10, 23, 16, 23, 40, 338_000_000, time.UTC)))
	assert.Equal(t, "docker compose build", sample.Command)
	assert.Equal(t, int64(2789), sample.DurationMs)
	assert.True(t, sample.Failed)
	assert.Equal(t, "build", sample.Job)
}

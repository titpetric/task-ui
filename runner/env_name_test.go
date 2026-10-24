package runner_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/runner"
)

// A name says per mode whether a run gets it; all passes every one.
func TestEnvName_Forwarded(t *testing.T) {
	t.Parallel()
	key := runner.EnvName{Name: "WORKSPACE_KEY", Sensitive: true}
	assert.False(t, key.Forwarded(runner.EnvNone))
	assert.False(t, key.Forwarded(runner.EnvAuto))
	assert.True(t, key.Forwarded(runner.EnvAll))
	path := runner.EnvName{Name: "PATH", None: true, Auto: true}
	assert.True(t, path.Forwarded(runner.EnvNone))
	assert.True(t, path.Forwarded(""))
}

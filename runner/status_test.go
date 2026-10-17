package runner_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
)

func TestStatus_Finished(t *testing.T) {
	t.Parallel()
	assert.False(t, runner.Status{State: model.StateRunning}.Finished())
	assert.True(t, runner.Status{State: model.StateFailed, ExitCode: 2}.Finished())
}

func TestEvent(t *testing.T) {
	t.Parallel()
	status := runner.Status{State: model.StateOK}
	event := runner.Event{Output: []byte("x"), Status: &status}
	assert.Equal(t, "x", string(event.Output))
	assert.True(t, event.Status.Finished())
}

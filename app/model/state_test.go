package model_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestStates(t *testing.T) {
	t.Parallel()
	states := model.States()
	assert.Len(t, states, 6)
	assert.Equal(t, model.StateQueued, states[0])
	assert.Equal(t, model.StateRunning, states[1])
}

func TestFinished(t *testing.T) {
	t.Parallel()
	assert.False(t, model.Finished(model.StateRunning))
	assert.False(t, model.Finished(model.StateQueued))
	assert.False(t, model.Finished(""))
	for _, state := range []model.State{model.StateOK, model.StateFailed, model.StateStopped, model.StateKilled} {
		assert.True(t, model.Finished(state), state)
	}
}

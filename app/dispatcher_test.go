package app

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
)

// NewDispatcher keeps the board, the users and the store it is given.
func TestNewDispatcher(t *testing.T) {
	t.Parallel()
	dispatcher := NewDispatcher(nil, nil, nil)
	assert.NotNil(t, dispatcher)
	assert.Nil(t, dispatcher.Board)
}

// Dispatch starts nothing on an instance with no board.
func TestDispatcher_Dispatch(t *testing.T) {
	t.Parallel()
	_, err := (&Dispatcher{}).Dispatch(context.Background(), "", "", "", "")
	assert.ErrorContains(t, err, "starts no module runs")
}

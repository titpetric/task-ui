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

// A module's prompt records the page its action names.
func TestOrigin(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "search", origin("/x/app", "prompts/search.md"))
	assert.Equal(t, "module:refinement", origin("/x/modules/refinement/", "prompts/refinement.md"))
}

// Dispatch starts nothing on an instance with no board.
func TestDispatcher_Dispatch(t *testing.T) {
	t.Parallel()
	_, err := (&Dispatcher{}).Dispatch(context.Background(), "", "", "", "")
	assert.ErrorContains(t, err, "starts no module runs")
}

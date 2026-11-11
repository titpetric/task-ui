package note_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/feature/note"
	"github.com/titpetric/task-ui/internal/assert"
)

// No runs ask nothing of the store, and have no summaries.
func TestBoard_Summaries(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 0, len((&note.Board{}).Summaries(context.Background(), nil)))
}

package note_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/feature/note"
	"github.com/titpetric/task-ui/internal/assert"
)

// A record without an author is refused before any storage is read.
func TestBoard_Record(t *testing.T) {
	t.Parallel()
	_, err := (&note.Board{}).Record(context.Background(), nil, "", "hello", "")
	assert.ErrorIs(t, err, note.ErrInvalid)
}

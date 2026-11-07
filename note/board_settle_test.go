package note

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// The last run of a thread to end settles its note. The ones that ended
// earlier leave it working. A failed run beside one that ended well
// settles the note failed, either order: a done would hide the failure.
// A thread under no count settles at once.
func TestBoardSettling(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name  string
		ended []string
		want  string
	}{
		{"both ended well", []string{model.NoteDone, model.NoteDone}, model.NoteDone},
		{"the first failed", []string{model.NoteFailed, model.NoteDone}, model.NoteFailed},
		{"the last failed", []string{model.NoteDone, model.NoteFailed}, model.NoteFailed},
		{"both failed", []string{model.NoteFailed, model.NoteFailed}, model.NoteFailed},
		{"three agents", []string{model.NoteDone, model.NoteFailed, model.NoteDone}, model.NoteFailed},
	} {
		board := &Board{}
		board.dispatching("n1", len(tc.ended))
		last := len(tc.ended) - 1
		for _, state := range tc.ended[:last] {
			assert.Equal(t, model.NoteWorking, board.settling("n1", state), tc.name)
		}
		assert.Equal(t, tc.want, board.settling("n1", tc.ended[last]), tc.name)
	}

	board := &Board{}
	assert.Equal(t, model.NoteDone, board.settling("n2", model.NoteDone), "no count stands")
	board.dispatching("n3", 1)
	assert.Equal(t, model.NoteDone, board.settling("n3", model.NoteDone), "one run alone settles the note")
	assert.Equal(t, model.NoteFailed, board.settling("n3", model.NoteFailed), "the count went with the last run")
	// A second message of a thread adds its run to the ones going, and
	// two messages on one thread settle it once.
	board.dispatching("n4", 1)
	board.dispatching("n4", 1)
	assert.Equal(t, model.NoteWorking, board.settling("n4", model.NoteDone))
	assert.Equal(t, model.NoteDone, board.settling("n4", model.NoteDone))
	board.dispatching("n5", 0)
	assert.Equal(t, model.NoteDone, board.settling("n5", model.NoteDone), "no agent, no count")
}

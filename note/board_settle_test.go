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

// The runs of one thread write one run on its note. The first to ask
// takes it; every run beside it leaves it there, so the note links to
// one run rather than to whichever wrote last. A thread whose runs all
// ended hands out the next one, and a run the board counts for no
// thread, a prompt's, takes its own.
func TestBoardClaims(t *testing.T) {
	t.Parallel()
	board := &Board{}
	board.dispatching("n1", 2)
	assert.Equal(t, "r1", board.claims("n1", "r1"))
	assert.Equal(t, "r1", board.claims("n1", "r2"), "the second run leaves the note's run alone")
	assert.Equal(t, model.NoteWorking, board.settling("n1", model.NoteDone))
	assert.Equal(t, model.NoteDone, board.settling("n1", model.NoteDone))

	board.dispatching("n1", 1)
	assert.Equal(t, "r3", board.claims("n1", "r3"), "a message after the runs ended")
	assert.Equal(t, "r4", board.claims("n2", "r4"), "no count, no claim")
}

// A notice on a thread comes from every agent working on it, a reply
// each, so a restart that cut two runs reads as two. A thread naming no
// agent gets the one reply its answerer leaves, under no name.
func TestNoticeAuthors(t *testing.T) {
	t.Parallel()
	assert.Equal(t, []string{"claude", "consuela"}, authors([]string{"claude", "consuela"}))
	assert.Equal(t, []string{"claude"}, authors([]string{"claude"}))
	assert.Equal(t, []string{""}, authors(nil), "the thread's answerer")
}

// A done reported over MCP while another run of the thread goes leaves
// the note working; the last run to end settles it on what was reported.
// A reported failure outlives a run that ended well, a working report
// folds nothing in, and a thread whose runs the board counts none of
// takes the report as it comes.
func TestBoardReported(t *testing.T) {
	t.Parallel()
	board := &Board{}
	board.dispatching("n1", 2)
	assert.Equal(t, model.NoteWorking, board.reported("n1", model.NoteDone), "the second agent is still working")
	assert.Equal(t, model.NoteWorking, board.settling("n1", model.NoteDone))
	assert.Equal(t, model.NoteDone, board.settling("n1", model.NoteDone), "the report settles with the last run")

	board.dispatching("n2", 1)
	assert.Equal(t, model.NoteWorking, board.reported("n2", model.NoteFailed))
	assert.Equal(t, model.NoteFailed, board.settling("n2", model.NoteDone), "a done would hide the reported failure")

	board.dispatching("n3", 1)
	assert.Equal(t, model.NoteWorking, board.reported("n3", model.NoteWorking), "a working report is no settle")
	assert.Equal(t, model.NoteDone, board.settling("n3", model.NoteDone), "it folded nothing in")

	assert.Equal(t, model.NoteDone, board.reported("n4", model.NoteDone), "no run of the thread goes")
}

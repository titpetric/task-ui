//go:build integration

package note_test

import (
	"context"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/app/model"
)

// A message that names two agents gets a run each. The first to end
// leaves the note working; the last settles it. Both failed here, so it
// reads failed.
func TestBoard_PostSettles(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	board, notes, people := newBoard(t)

	states := watch(board)
	thread, err := board.Post(ctx, people["ops"], "", "@claude take the API, @consuela take the docs")
	assert.NoError(t, err)
	assert.Equal(t, []string{"claude", "consuela"}, answers(t, notes, thread.ID))
	assert.Equal(t, []string{model.NoteOpen, model.NoteWorking, model.NoteFailed}, states(thread.ID),
		"the note waits for the second agent")

	ended, err := notes.Get(ctx, thread.ID)
	assert.NoError(t, err)
	assert.Equal(t, model.NoteFailed, ended.State, "the last run of the thread settles it")
}

// A reply that names a second agent counts that agent's run too. The
// note settles once both ended.
func TestBoard_ReplySettles(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	board, notes, people := newBoard(t)

	thread, err := board.Post(ctx, people["ops"], "", "@claude take the API")
	assert.NoError(t, err)

	states := watch(board)
	_, err = board.Post(ctx, people["ops"], thread.ID, "@claude and @consuela, both of you")
	assert.NoError(t, err)
	assert.Equal(t, []string{model.NoteOpen, model.NoteWorking, model.NoteFailed}, states(thread.ID),
		"the reply opens the note again and both agents run")
	assert.Equal(t, []string{"claude", "ops", "claude", "consuela"}, answers(t, notes, thread.ID))
}

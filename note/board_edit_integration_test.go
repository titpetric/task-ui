//go:build integration

package note_test

import (
	"context"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
)

// An edited instruction runs again for each agent it addresses, and each
// of them rewrites the reply it wrote, leaving the other agent's alone.
// The note waits for the last of the runs the edit started.
func TestBoard_Edit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	board, notes, people := newBoard(t)

	thread := &model.Note{UserName: "ops", Body: "@claude take the API, @consuela take the docs", State: model.NoteOpen}
	assert.NoError(t, notes.Create(ctx, thread))
	assert.NoError(t, notes.Create(ctx, &model.Note{ParentID: thread.ID, UserName: "claude", Body: "the API is done"}))
	assert.NoError(t, notes.Create(ctx, &model.Note{ParentID: thread.ID, UserName: "consuela", Body: "the docs are done"}))

	states := watch(board)
	_, err := board.Edit(ctx, people["ops"], thread.ID, "@claude take the API, @consuela take the docs, in English")
	assert.NoError(t, err)
	// The edit publishes the note, then opens it again: two reads of open.
	assert.Equal(t, []string{model.NoteOpen, model.NoteOpen, model.NoteWorking, model.NoteFailed}, states(thread.ID))

	replies, err := notes.Replies(ctx, []string{thread.ID})
	assert.NoError(t, err)
	assert.Len(t, replies[thread.ID], 2, "each agent rewrote its answer")
	assert.Equal(t, []string{"claude", "consuela"}, answers(t, notes, thread.ID))
	for _, reply := range replies[thread.ID] {
		assert.Contains(t, reply.Body, noRoot, reply.UserName)
	}
}

// Every agent a thread names answers it. The second agent's reply is an
// answer as the first one's is: its author rewrites neither where it
// stands, and no admin either.
func TestBoard_EditAnswer(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	board, notes, people := newBoard(t)

	thread := &model.Note{UserName: "ops", Body: "@claude take the API, @consuela take the docs", State: model.NoteWorking}
	assert.NoError(t, notes.Create(ctx, thread))
	first := &model.Note{ParentID: thread.ID, UserName: "claude", Body: "the API is done"}
	assert.NoError(t, notes.Create(ctx, first))
	second := &model.Note{ParentID: thread.ID, UserName: "consuela", Body: "the docs are done"}
	assert.NoError(t, notes.Create(ctx, second))
	mine := &model.Note{ParentID: thread.ID, UserName: "ops", Body: "thanks"}
	assert.NoError(t, notes.Create(ctx, mine))

	_, err := board.Edit(ctx, people["claude"], first.ID, "the API is done, and tested")
	assert.ErrorIs(t, err, note.ErrInvalid, "the first agent's answer stands")
	_, err = board.Edit(ctx, people["consuela"], second.ID, "the docs are done, and spelled")
	assert.ErrorIs(t, err, note.ErrInvalid, "the second agent's answer stands too")
	_, err = board.Edit(ctx, people["ops"], second.ID, "the docs are done, and spelled")
	assert.ErrorIs(t, err, note.ErrInvalid, "an admin rewrites no answer either")

	held, err := notes.Get(ctx, second.ID)
	assert.NoError(t, err)
	assert.Equal(t, "the docs are done", held.Body)

	_, err = board.Edit(ctx, people["ops"], mine.ID, "thanks, both")
	assert.NoError(t, err, "a reply of the asker's own is theirs to rewrite")
}

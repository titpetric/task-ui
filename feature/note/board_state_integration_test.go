//go:build integration

package note_test

import (
	"context"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/assert"
)

// A restart starts every agent an open instruction with no run names, one
// run each, and leaves a thread that names nobody alone. Its runs count
// as a message's do, and the note waits for the last.
func TestBoard_Recover(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	board, notes, _ := newBoard(t)

	thread := &model.Note{UserName: "ops", Body: "@claude take the API, @consuela take the docs", State: model.NoteOpen}
	assert.NoError(t, notes.Create(ctx, thread))
	message := &model.Note{UserName: "ops", Body: "the archive is big", State: model.NoteOpen}
	assert.NoError(t, notes.Create(ctx, message))

	states := watch(board)
	assert.NoError(t, board.Recover(ctx))
	assert.Equal(t, []string{"claude", "consuela"}, answers(t, notes, thread.ID))
	assert.Len(t, answers(t, notes, message.ID), 0)
	assert.Equal(t, []string{model.NoteWorking, model.NoteFailed}, states(thread.ID))

	ended, err := notes.Get(ctx, thread.ID)
	assert.NoError(t, err)
	assert.Equal(t, model.NoteFailed, ended.State, "the board has no root to run in")
}

// A restart drops the empty answer a crash left before its run started, and
// asks the open prompt again.
func TestBoard_RecoverPlaceholder(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	board, notes, people := newBoard(t)

	prompt := &model.Note{UserName: "ops", UserID: people["ops"].ID, Body: "why is the build red", State: model.NoteOpen, WorkspaceID: model.Default, Dir: "infra"}
	assert.NoError(t, notes.Create(ctx, prompt))
	left := &model.Note{ParentID: prompt.ID, UserName: "claude", Agent: "claude", State: model.NoteWorking, RunID: model.NewID()}
	assert.NoError(t, notes.CreateAnswer(ctx, left))

	assert.NoError(t, board.Recover(ctx))
	replies, err := notes.Replies(ctx, []string{prompt.ID})
	assert.NoError(t, err)
	assert.Len(t, replies[prompt.ID], 1, "the crash's empty answer is gone")
	assert.NotEqual(t, left.ID, replies[prompt.ID][0].ID)
}

// The restart notice comes from every agent of the thread, a reply each:
// a thread two agents worked on reads as two cut runs, not one. The note
// settles once.
func TestBoard_RecoverNotice(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	board, notes, _ := newBoard(t)

	thread := &model.Note{UserName: "ops", Body: "@claude take the API, @consuela take the docs", State: model.NoteWorking, RunID: "r1"}
	assert.NoError(t, notes.Create(ctx, thread))
	alone := &model.Note{UserName: "ops", Body: "@claude take the API", State: model.NoteWorking, RunID: "r2"}
	assert.NoError(t, notes.Create(ctx, alone))

	states := watch(board)
	assert.NoError(t, board.Recover(ctx))
	assert.Equal(t, []string{"claude", "consuela"}, answers(t, notes, thread.ID))
	assert.Equal(t, []string{"claude"}, answers(t, notes, alone.ID), "one agent, one notice")
	assert.Equal(t, []string{model.NoteFailed}, states(thread.ID), "the note settles once")

	ended, err := notes.Get(ctx, thread.ID)
	assert.NoError(t, err)
	assert.Equal(t, model.NoteFailed, ended.State)
}

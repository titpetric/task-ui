//go:build integration

package note_test

import (
	"context"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// A restart starts every agent an open instruction with no run names, one
// run each, and leaves a thread that names nobody alone.
func TestBoard_Recover(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	board, notes, _ := newBoard(t)

	thread := &model.Note{UserName: "ops", Body: "@claude take the API, @consuela take the docs", State: model.NoteOpen}
	assert.NoError(t, notes.Create(ctx, thread))
	message := &model.Note{UserName: "ops", Body: "the archive is big", State: model.NoteOpen}
	assert.NoError(t, notes.Create(ctx, message))

	assert.NoError(t, board.Recover(ctx))
	assert.Equal(t, []string{"claude", "consuela"}, answers(t, notes, thread.ID))
	assert.Len(t, answers(t, notes, message.ID), 0)

	ended, err := notes.Get(ctx, thread.ID)
	assert.NoError(t, err)
	assert.Equal(t, model.NoteFailed, ended.State, "the board has no root to run in")
}

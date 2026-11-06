//go:build integration

package note_test

import (
	"context"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// An edited instruction runs again for each agent it addresses, and each
// of them rewrites the reply it wrote, leaving the other agent's alone.
func TestBoard_Edit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	board, notes, people := newBoard(t)

	thread := &model.Note{UserName: "ops", Body: "@claude take the API, @consuela take the docs", State: model.NoteOpen}
	assert.NoError(t, notes.Create(ctx, thread))
	assert.NoError(t, notes.Create(ctx, &model.Note{ParentID: thread.ID, UserName: "claude", Body: "the API is done"}))
	assert.NoError(t, notes.Create(ctx, &model.Note{ParentID: thread.ID, UserName: "consuela", Body: "the docs are done"}))

	_, err := board.Edit(ctx, people["ops"], thread.ID, "@claude take the API, @consuela take the docs, in English")
	assert.NoError(t, err)

	replies, err := notes.Replies(ctx, []string{thread.ID})
	assert.NoError(t, err)
	assert.Len(t, replies[thread.ID], 2, "each agent rewrote its answer")
	assert.Equal(t, []string{"claude", "consuela"}, answers(t, notes, thread.ID))
	for _, reply := range replies[thread.ID] {
		assert.Contains(t, reply.Body, noRoot, reply.UserName)
	}
}

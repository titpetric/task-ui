//go:build integration

package app_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// A module records the note its dispatch answered with and asks for the
// run when it draws the page, which is later: the note holds the run
// once the board has started one. A note with no run yet is a label
// with nothing in it, not an error.
func TestModuleRuns_RunOf(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := storage.OpenAndMigrate(ctx, ":memory:")
	assert.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, db.Close()) })

	notes := storage.NewNoteStorage(db)
	runs := app.NewModuleRuns(notes)

	waiting := &model.Note{Channel: model.ChannelNotes, UserName: "miguel", Body: "refine it"}
	assert.NoError(t, notes.Create(ctx, waiting))
	assert.Equal(t, "", runs.RunOf(ctx, waiting.ID), "a note whose run has not started")

	// The board sets the run on the note when it starts one, which is
	// what makes the label appear on the next page load.
	assert.NoError(t, notes.SetState(ctx, waiting.ID, model.NoteOpen, "06grxxqjec8jjt19vd65wfdg8r"))
	assert.Equal(t, "06grxxqjec8jjt19vd65wfdg8r", runs.RunOf(ctx, waiting.ID))

	assert.Equal(t, "", runs.RunOf(ctx, "01NOSUCHNOTE00000000000000"), "a note that is not there")
	assert.Equal(t, "", app.NewModuleRuns(nil).RunOf(ctx, waiting.ID), "an instance that keeps no notes")
}

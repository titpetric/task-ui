package note

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/assert"
)

// A prompt names its place by the area, the run it is about or its folder.
func TestBoardPlace(t *testing.T) {
	t.Parallel()
	board := &Board{}
	ctx := context.Background()

	assert.Equal(t, "in "+model.PlaceLabel("infra/app"), board.place(ctx, &model.Note{Dir: "infra/app"}))
	assert.Equal(t, "about the run r1, in "+model.PlaceLabel("infra/app"), board.place(ctx, &model.Note{Dir: "infra/app", AboutRunID: "r1"}))
	assert.Equal(t, "the prompt of the Workflows page, in "+model.PlaceLabel("infra/app"), board.place(ctx, &model.Note{Dir: "infra/app", Area: model.AreaWorkflows}))
}

// A thread's link is the page of its run, area, workflow run or folder.
func TestBoardThreadPath(t *testing.T) {
	t.Parallel()
	board := &Board{}
	ctx := context.Background()

	assert.Equal(t, "/runs/r1#note-n1", board.threadPath(ctx, &model.Note{ID: "n1", AboutRunID: "r1"}))
	assert.Equal(t, "/workflows#note-n1", board.threadPath(ctx, &model.Note{ID: "n1", Area: model.AreaWorkflows}))
	assert.Equal(t, "/f/internal#note-n1", board.threadPath(ctx, &model.Note{ID: "n1"}))
	assert.Equal(t, "/f/ws/src#note-n1", board.threadPath(ctx, &model.Note{ID: "n1", Dir: "ws/src", WorkspaceID: "ws"}))
	assert.Equal(t, "/f/internal/infra/app#note-n1", board.threadPath(ctx, &model.Note{ID: "n1", Dir: "infra/app"}))
}

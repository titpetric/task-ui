package note_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
)

// No run has no report, nor has a run of Claude's still at work.
func TestBoard_Report(t *testing.T) {
	t.Parallel()
	board := &note.Board{}
	assert.Nil(t, board.Report(context.Background(), nil))
	assert.Nil(t, board.Report(context.Background(), &model.Run{ID: "c1", Job: model.PromptJob, State: model.StateRunning}))
}

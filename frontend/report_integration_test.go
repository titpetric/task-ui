//go:build integration

package frontend_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
)

// A run's report is Claude's summary of it, as a document, for a user
// with prompt/view; a run without one, or a user without it, has none.
func TestReport(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	run := f.run(t, "infra", "plan", "ops", model.StateFailed)
	path := "/runs/" + run.ID + "/report"
	assert.Equal(t, http.StatusNotFound, f.get(t, asAdmin, path).Status, "no summary yet")
	assert.Equal(t, http.StatusNotFound, f.get(t, asAdmin, "/runs/missing/report").Status)

	summary := f.run(t, "infra", model.PromptJob, "ops", model.StateOK)
	thread := &model.Note{UserName: "ops", Body: note.SummaryQuestion, State: model.NoteDone, WorkspaceID: "ws", Dir: "infra", AboutRunID: run.ID}
	assert.NoError(t, f.opts.Board.Notes.Create(ctx, thread))
	assert.NoError(t, f.opts.Board.Notes.Create(ctx, &model.Note{ParentID: thread.ID, UserName: "claude", Body: "The **plan** failed on a lock.", RunID: summary.ID, WorkspaceID: "ws", Dir: "infra"}))

	page := f.page(t, asAdmin, path)
	assert.Contains(t, page, "Claude's summary of the run")
	assert.Contains(t, page, "<strong>plan</strong> failed on a lock.")
	assert.Contains(t, page, `href="/runs/`+run.ID+`">Run</a>`)
	assert.Equal(t, http.StatusNotFound, f.get(t, asOps, path).Status, "no prompt/view")
}

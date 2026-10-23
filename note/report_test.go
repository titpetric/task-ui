package note_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
)

// A report page is under its run's page.
func TestReport_Href(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "/runs/r1/report", (&note.Report{RunID: "r1"}).Href())
}

// A summary thread reports its first answer, pending until there is
// one, and nothing once it failed without one; another thread is no
// summary.
func TestSummaryOf(t *testing.T) {
	t.Parallel()
	thread := &model.Note{ID: "t", Body: note.SummaryQuestion, AboutRunID: "r1", UserName: "ops", RunID: "c1", State: model.NoteWorking}
	pending := note.SummaryOf(thread, nil)
	assert.True(t, pending.Pending)
	assert.Equal(t, "r1", pending.RunID)
	assert.Equal(t, "ops", pending.By)

	done := *thread
	done.RunID, done.State = "", model.NoteDone
	report := note.SummaryOf(&done, []model.Note{{ID: "a", RunID: "c1", Body: "## Passed"}, {ID: "a2", RunID: "c2", Body: "later"}})
	assert.False(t, report.Pending)
	assert.Equal(t, "## Passed", report.Body)
	assert.Equal(t, "t", report.ThreadID)

	failed := done
	failed.State = model.NoteFailed
	assert.Nil(t, note.SummaryOf(&failed, nil))
	assert.Nil(t, note.SummaryOf(&model.Note{ID: "x", Body: "why?", AboutRunID: "r1"}, nil))
}

// A summary kept without a thread, or a thread that reads a summary's
// question, is a summary; another question has an answer.
func TestReport_IsSummary(t *testing.T) {
	t.Parallel()
	assert.True(t, (&note.Report{}).IsSummary())
	assert.True(t, (&note.Report{ThreadID: "t", Question: note.SummaryQuestion}).IsSummary())
	assert.True(t, (&note.Report{ThreadID: "t", Question: note.BatchSummaryQuestion}).IsSummary())
	assert.False(t, (&note.Report{ThreadID: "t", Question: "why did it fail?"}).IsSummary())
}

package note_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
)

// A summary is the thread Run and summarize asks about a run; its run
// is the first answer's.
func TestIsSummary(t *testing.T) {
	t.Parallel()
	assert.True(t, note.IsSummary(&model.Note{AboutRunID: "r1", Body: note.SummaryQuestion}))
	assert.True(t, note.IsSummary(&model.Note{AboutRunID: "r1", Body: note.BatchSummaryQuestion}))
	assert.False(t, note.IsSummary(&model.Note{AboutRunID: "r1", Body: "why did it fail"}))
	assert.False(t, note.IsSummary(&model.Note{Body: note.SummaryQuestion}), "about no run")
	assert.False(t, note.IsSummary(&model.Note{AboutRunID: "r1", ParentID: "n1", Body: note.SummaryQuestion}), "a reply")
	assert.False(t, note.IsSummary(nil))
}

func TestSummaryRunID(t *testing.T) {
	t.Parallel()
	thread := &model.Note{RunID: "r3"}
	assert.Equal(t, "r3", note.SummaryRunID(thread, nil), "while Claude works")
	replies := []model.Note{{Body: "a user's reply"}, {RunID: "r2"}, {RunID: "r3"}}
	assert.Equal(t, "r2", note.SummaryRunID(thread, replies), "the first answer")
	assert.Equal(t, "", note.SummaryRunID(nil, nil))
}

package note

import (
	"testing"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/assert"
)

// TestConversationRun asserts events from one run stay in one group.
func TestConversationRun(t *testing.T) {
	events := []model.NoteEvent{
		{QuestionID: "q", RunID: "r", Seq: 1, Kind: "text"},
		{QuestionID: "q", RunID: "r", Seq: 2, Kind: "result"},
		{QuestionID: "q2", RunID: "r2", Seq: 1, Kind: "text"},
	}
	runs := groupEvents(events)
	assert.Len(t, runs, 2)
	assert.Equal(t, "r", runs[0].Run)
	assert.Len(t, runs[0].Events, 2)
	assert.Equal(t, "r2", runs[1].Run)
	assert.Len(t, runs[1].Events, 1)
}

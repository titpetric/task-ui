//go:build integration

package storage_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/storage"
)

// A thread's usage sums what every run read and wrote, and counts its tool calls.
func TestNoteEventStorage_Usage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	events := storage.NewNoteEventStorage(newDB(t))

	for _, event := range []*model.NoteEvent{
		{NoteID: "n1", RunID: "r1", QuestionID: "q1", Seq: 1, Kind: "tool_use", Data: `{"name":"Bash"}`},
		{NoteID: "n1", RunID: "r1", QuestionID: "q1", Seq: 2, Kind: "result", Data: `{"input_tokens":100,"cache_read_input_tokens":50,"cost_usd":0.25}`},
		{NoteID: "n1", RunID: "r2", QuestionID: "q1", Seq: 1, Kind: "result", Data: `{"input_tokens":10,"output_tokens":5,"cost_usd":0.05}`},
		{NoteID: "n2", RunID: "r3", QuestionID: "q3", Seq: 1, Kind: "text", Body: "hi"},
	} {
		assert.NoError(t, events.Append(ctx, event))
	}

	usage, err := events.Usage(ctx, []string{"n1", "n2"})
	assert.NoError(t, err)
	assert.Equal(t, 1, usage["n1"].Tools)
	assert.Equal(t, 160, usage["n1"].Tokens)
	assert.Equal(t, 0.3, usage["n1"].Cost)
	assert.Equal(t, storage.ThreadUsage{}, usage["n2"])
}

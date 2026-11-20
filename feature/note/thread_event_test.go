package note

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/assert"
)

// TestThreadEvent asserts the payload carries the record and its JSON data.
func TestThreadEvent(t *testing.T) {
	data, err := json.Marshal(ThreadEvent{Note: "n", Run: "r", Question: "q", Seq: 2, Kind: "text", Body: "hi", Data: json.RawMessage(`{"k":1}`)})
	assert.NoError(t, err)
	assert.JSONEq(t, `{"note":"n","run":"r","question":"q","seq":2,"kind":"text","body":"hi","data":{"k":1}}`, string(data))
}

// TestEventOfMapsTheDisplayFields asserts a tool call reaches the pane under
// the names the template reads.
func TestEventOfMapsTheDisplayFields(t *testing.T) {
	row := &model.NoteEvent{ID: "e", NoteID: "n", RunID: "r", QuestionID: "q", Seq: 3, Kind: "tool_use", Body: "> Bash", Data: `{"name":"Bash","input":{"command":"go test ./..."}}`}
	event := EventOf(row)

	assert.Equal(t, "e", event.ID)
	assert.Equal(t, int64(3), event.Seq)
	assert.Equal(t, "Bash", event.Tool)
	assert.Equal(t, `{"command":"go test ./..."}`, event.Input)
}

// TestEventOfMarksAnErrorAndCuts asserts a tool result carries its error and
// is cut where the stored body is longer than a card.
func TestEventOfMarksAnErrorAndCuts(t *testing.T) {
	row := &model.NoteEvent{Kind: "tool_result", Body: strings.Repeat("x", threadCardCut+10), Data: `{"is_error":true}`}
	event := EventOf(row)

	assert.Equal(t, true, event.Error)
	assert.Equal(t, threadCardCut+len("\n... (truncated)"), len(event.Body))
}

// TestEventOfReadsAStatus asserts the status carries what Claude is doing.
func TestEventOfReadsAStatus(t *testing.T) {
	row := &model.NoteEvent{Kind: "status", Body: "Running Bash", Data: `{"state":"tool","tool":"Bash"}`}
	event := EventOf(row)

	assert.Equal(t, "tool", event.State)
	assert.Equal(t, "Bash", event.Tool)
}

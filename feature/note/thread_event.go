package note

import (
	"encoding/json"
	"time"

	"github.com/titpetric/task-ui/app/model"
)

// threadCardCut is how much of a tool's input or result a card shows.
const threadCardCut = 4000

// ThreadEvent is one item of a thread's live work, on its topic. It is also
// the shape the pane draws, so a reloaded transcript and a streamed event
// carry the same fields.
type ThreadEvent struct {
	ID       string          `json:"id,omitempty"`
	Note     string          `json:"note"`
	Run      string          `json:"run"`
	Question string          `json:"question"`
	Seq      int64           `json:"seq"`
	Kind     string          `json:"kind"`
	ToolID   string          `json:"tool_id,omitempty"`
	Tool     string          `json:"tool,omitempty"`
	Input    string          `json:"input,omitempty"`
	Error    bool            `json:"error,omitempty"`
	State    string          `json:"state,omitempty"`
	Body     string          `json:"body,omitempty"`
	Created  string          `json:"created_at,omitempty"`
	Data     json.RawMessage `json:"data,omitempty"`
}

// EventOf returns the display shape of a stored event, with the tool input
// and result cut the way the pane cuts a card.
func EventOf(event *model.NoteEvent) ThreadEvent {
	out := ThreadEvent{
		ID: event.ID, Note: event.NoteID, Run: event.RunID, Question: event.QuestionID,
		Seq: event.Seq, Kind: event.Kind, ToolID: event.ToolID, Body: event.Body,
	}
	if event.CreatedAt != nil {
		out.Created = event.CreatedAt.UTC().Format(time.RFC3339Nano)
	}
	var data struct {
		Name    string          `json:"name"`
		Input   json.RawMessage `json:"input"`
		IsError bool            `json:"is_error"`
		State   string          `json:"state"`
		Tool    string          `json:"tool"`
	}
	if err := json.Unmarshal([]byte(event.Data), &data); err != nil {
		return out
	}
	switch event.Kind {
	case "tool_use":
		out.Tool = data.Name
		out.Input = cardInput(data.Input)
	case "tool_result":
		out.Error = data.IsError
		out.Body = cutCard(out.Body)
	case "status":
		out.State = data.State
		out.Tool = data.Tool
	case "error":
		out.Error = true
	}
	return out
}

// cardInput is a tool's input as a card shows it: its JSON text, cut.
func cardInput(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return cutCard(string(raw))
	}
	text, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return cutCard(string(text))
}

// cutCard is text as a card holds it, marked where it was cut.
func cutCard(text string) string {
	if len(text) <= threadCardCut {
		return text
	}
	return text[:threadCardCut] + "\n... (truncated)"
}

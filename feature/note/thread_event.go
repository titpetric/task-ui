package note

import (
	"encoding/json"
)

// ThreadEvent is one item of a thread's live work, on its topic.
type ThreadEvent struct {
	Note     string          `json:"note"`
	Run      string          `json:"run"`
	Question string          `json:"question"`
	Seq      int64           `json:"seq"`
	Kind     string          `json:"kind"`
	ToolID   string          `json:"tool_id,omitempty"`
	Body     string          `json:"body,omitempty"`
	Data     json.RawMessage `json:"data,omitempty"`
}

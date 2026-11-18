package note

import (
	"encoding/json"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
)

// TestThreadEvent asserts the payload carries the record and its JSON data.
func TestThreadEvent(t *testing.T) {
	data, err := json.Marshal(ThreadEvent{Note: "n", Run: "r", Question: "q", Seq: 2, Kind: "text", Body: "hi", Data: json.RawMessage(`{"k":1}`)})
	assert.NoError(t, err)
	assert.JSONEq(t, `{"note":"n","run":"r","question":"q","seq":2,"kind":"text","body":"hi","data":{"k":1}}`, string(data))
}

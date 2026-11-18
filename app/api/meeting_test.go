package api

import (
	"encoding/json"
	"testing"

	"golang.org/x/net/websocket"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/assert"
)

// A frame keeps whether it came as text or binary, which Message.Receive drops.
func TestUnmarshalFrame(t *testing.T) {
	t.Parallel()
	var frame wireFrame
	assert.NoError(t, unmarshalFrame([]byte{1, 2}, websocket.BinaryFrame, &frame))
	assert.Equal(t, byte(websocket.BinaryFrame), frame.Kind)
	assert.Len(t, frame.Data, 2)
}

// A control frame names what it is, the text it holds and how long the speech lasted.
func TestPipeControl(t *testing.T) {
	t.Parallel()
	var control pipeControl
	assert.NoError(t, json.Unmarshal([]byte(`{"type":"line","text":"hello","duration_ms":1500}`), &control))
	assert.Equal(t, "line", control.Type)
	assert.Equal(t, "hello", control.Text)
	assert.Equal(t, int64(1500), control.DurationMS)
}

// A summary is the markdown and the same rendered for a page.
func TestMeetingSummaryJSON(t *testing.T) {
	t.Parallel()
	room := &model.Meeting{ID: "m", Summary: "## Topics discussed\n\n- the release", SummaryState: model.SummaryDone, SummaryUserName: "ana"}
	view := meetingSummaryJSON(room)
	assert.Equal(t, model.SummaryDone, view["state"])
	assert.Equal(t, "ana", view["user_name"])
	html, _ := view["html"].(string)
	assert.Contains(t, html, `<h2 id="topics-discussed">Topics discussed</h2>`)
}

package meeting

import (
	"encoding/json"
	"testing"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/assert"
)

// A text frame holds its message as JSON and no binary part.
func TestTextFrame(t *testing.T) {
	t.Parallel()
	line := &model.MeetingLine{ID: "l1", Kind: model.LineSaid, Body: "hello"}
	frame := TextFrame(Message{Type: "line", Line: line})
	var got Message
	assert.NoError(t, json.Unmarshal(frame.Text, &got))
	assert.Equal(t, "line", got.Type)
	assert.Equal(t, "hello", got.Line.Body)
	assert.Equal(t, 0, len(frame.Data))
}

// An audio frame names the speaker and holds the PCM bytes itself.
func TestAudioFrame(t *testing.T) {
	t.Parallel()
	frame := AudioFrame("u1", "ana", 3, 750, []byte{1, 2})
	var got Message
	assert.NoError(t, json.Unmarshal(frame.Text, &got))
	assert.Equal(t, "audio", got.Type)
	assert.Equal(t, "ana", got.UserName)
	assert.Equal(t, int64(3), got.Seq)
	assert.Equal(t, Rate, got.Rate)
	assert.Equal(t, []byte{1, 2}, frame.Data)
}

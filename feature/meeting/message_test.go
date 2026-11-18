package meeting

import (
	"testing"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/assert"
)

// A message names its frame and holds the fields that frame sends.
func TestMessage(t *testing.T) {
	t.Parallel()
	line := &model.MeetingLine{ID: "l1", Kind: model.LineSaid, Body: "hello"}
	msg := Message{Type: "line", Line: line, UserID: "u1", UserName: "ana", Seq: 2, Rate: Rate}
	assert.Equal(t, "line", msg.Type)
	assert.Equal(t, "hello", msg.Line.Body)
	assert.Equal(t, "ana", msg.UserName)
}

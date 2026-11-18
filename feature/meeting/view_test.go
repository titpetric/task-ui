package meeting

import (
	"testing"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/assert"
)

// A view groups a meeting with its participants and its log.
func TestView(t *testing.T) {
	t.Parallel()
	view := &View{
		Meeting:      &model.Meeting{ID: "m", Title: "Standup"},
		Participants: []model.MeetingParticipant{{UserID: "u1", UserName: "ana"}},
		Lines:        []model.MeetingLine{{ID: "l1", Body: "hello"}},
	}
	assert.Equal(t, "Standup", view.Meeting.Title)
	assert.Len(t, view.Participants, 1)
	assert.Equal(t, "hello", view.Lines[0].Body)
}

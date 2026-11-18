package model

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
)

// A meeting holds the room's state and summary fields.
func TestMeeting(t *testing.T) {
	t.Parallel()
	m := Meeting{ID: "m", Title: "Standup", State: MeetingOpen, LimitSeconds: MeetingLimit}
	assert.Equal(t, "Standup", m.Title)
	assert.Equal(t, MeetingOpen, m.State)
	assert.Equal(t, 3600, MeetingLimit)
}

// A participant holds the user, their last sighting and how often they joined.
func TestMeetingParticipant(t *testing.T) {
	t.Parallel()
	p := MeetingParticipant{MeetingID: "m", UserID: "u", UserName: "ana", Joins: 2}
	assert.Equal(t, "ana", p.UserName)
	assert.Equal(t, int64(2), p.Joins)
}

// A line holds its kind and the span of the speaker's channel it covers.
func TestMeetingLine(t *testing.T) {
	t.Parallel()
	l := MeetingLine{MeetingID: "m", Kind: LineSaid, Body: "hello", StartMs: 10, EndMs: 20}
	assert.Equal(t, LineSaid, l.Kind)
	assert.Equal(t, int64(20), l.EndMs)
	assert.Equal(t, "join", LineJoin)
}

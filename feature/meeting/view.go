package meeting

import (
	"github.com/titpetric/task-ui/app/model"
)

// View is a room as its page draws it: the meeting, the participants and the log so far.
type View struct {
	Meeting      *model.Meeting
	Participants []model.MeetingParticipant
	Lines        []model.MeetingLine
}

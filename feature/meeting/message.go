package meeting

import (
	"github.com/titpetric/task-ui/app/model"
)

// Message holds one frame of the meeting pipe: control, transcript, or the header of a binary audio frame.
type Message struct {
	Type         string                     `json:"type"`
	Meeting      *model.Meeting             `json:"meeting,omitempty"`
	Participants []model.MeetingParticipant `json:"participants,omitempty"`
	Lines        []model.MeetingLine        `json:"lines,omitempty"`
	Line         *model.MeetingLine         `json:"line,omitempty"`
	UserID       string                     `json:"user_id,omitempty"`
	UserName     string                     `json:"user_name,omitempty"`
	Text         string                     `json:"text,omitempty"`
	Muted        bool                       `json:"muted,omitempty"`
	DurationMS   int64                      `json:"duration_ms,omitempty"`
	Seq          int64                      `json:"seq,omitempty"`
	Rate         int                        `json:"rate,omitempty"`
	StartMS      int64                      `json:"start_ms,omitempty"`
	State        string                     `json:"state,omitempty"`
	Message      string                     `json:"message,omitempty"`
	HTML         string                     `json:"html,omitempty"`
}

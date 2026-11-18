package note

import (
	"github.com/titpetric/task-ui/app/model"
)

// ConversationRun groups one run's events by the question that started it.
type ConversationRun struct {
	Question string            `json:"question"`
	Run      string            `json:"run"`
	Answer   string            `json:"answer"`
	Events   []model.NoteEvent `json:"events"`
}

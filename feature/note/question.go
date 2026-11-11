package note

import (
	"github.com/titpetric/task-ui/model"
)

// Question returns the note an answer replies to: the reply before it
// in the thread, or the thread.
func Question(thread *model.Note, replies []model.Note, answer *model.Note) *model.Note {
	for i := range replies {
		if replies[i].ID == answer.ID {
			if i > 0 {
				return &replies[i-1]
			}
			break
		}
	}
	return thread
}

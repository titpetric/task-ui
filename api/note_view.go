package api

import (
	"github.com/titpetric/task-ui/model"
)

// NoteView is a note with its replies, oldest first.
type NoteView struct {
	model.Note
	Replies []model.Note `json:"replies"`
}

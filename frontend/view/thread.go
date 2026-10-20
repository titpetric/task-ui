package view

import (
	"github.com/titpetric/task-ui/model"
)

// Thread is a note with its replies, oldest first.
type Thread struct {
	model.Note
	Replies []model.Note
}

// Initials are the author's, for the avatar.
func (t Thread) Initials() string {
	return (&model.User{Name: t.UserName}).Initials()
}

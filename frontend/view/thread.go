package view

import (
	"github.com/titpetric/task-ui/model"
)

// Thread is a note with its replies, oldest first, and Answerers, the
// user names of the agents that answer it.
type Thread struct {
	model.Note
	Replies   []model.Note
	Answerers []string
}

// Initials are the author's, for the avatar.
func (t Thread) Initials() string {
	return (&model.User{Name: t.UserName}).Initials()
}

// Answer reports whether a reply is an agent's answer, which changes only when its question runs again.
func (t Thread) Answer(reply model.Note) bool {
	if t.State == "" {
		return false
	}
	for _, name := range t.Answerers {
		if reply.UserName == name {
			return true
		}
	}
	return false
}


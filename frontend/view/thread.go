package view

import (
	"github.com/titpetric/task-ui/model"
)

// Thread is a note with its replies, oldest first, and Answerer, the
// user name the agent's answers in it carry.
type Thread struct {
	model.Note
	Replies  []model.Note
	Answerer string
}

// Initials are the author's, for the avatar.
func (t Thread) Initials() string {
	return (&model.User{Name: t.UserName}).Initials()
}

// Answer reports whether a reply is the agent's answer to the thread's
// instruction or prompt, which changes only when its question runs again.
func (t Thread) Answer(reply model.Note) bool {
	return t.State != "" && t.Answerer != "" && reply.UserName == t.Answerer
}

// Edits reports whether the user edits a note of the thread, the thread's
// own or a reply: its author or an admin, the agent's answers excepted.
func (t Thread) Edits(user *model.User, note model.Note) bool {
	if user == nil || (note.ID != t.ID && t.Answer(note)) {
		return false
	}
	return note.UserName == user.Name || user.CanAdmin()
}

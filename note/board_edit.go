package note

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
)

// ErrNotAuthor reports an edit of a note by someone who is neither its
// author nor an admin.
var ErrNotAuthor = errors.New("only its author or an admin edits a note")

// Edit puts a new body in a note or a reply, its author's or any for an
// admin, and marks it edited. An instruction runs again on the edited
// text: a prompt, a reply in a prompt's thread, or a note or a reply on
// the board that names the agent. The agent's answer to it, the
// first answer after it in the thread, shows working and then takes the
// new answer in place, from a run under the same run ID; with no answer
// yet, the agent replies as it does to a new note. The agent's own
// answers change only that way.
func (b *Board) Edit(ctx context.Context, user *model.User, id, body string) (*model.Note, error) {
	if user == nil || strings.TrimSpace(body) == "" {
		return nil, fmt.Errorf("%w: a note needs an author and some text", ErrInvalid)
	}
	note, err := b.Notes.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if note.UserName != user.Name && !user.CanAdmin() {
		return nil, ErrNotAuthor
	}
	thread := note
	if note.ParentID != "" {
		if thread, err = b.Notes.Get(ctx, note.ParentID); err != nil {
			return nil, err
		}
		if thread.State != "" && note.UserName == b.Answerer() {
			return nil, fmt.Errorf("%w: the agent's answer changes when its question runs again", ErrInvalid)
		}
	}
	var req runner.Request
	if thread.IsPrompt() {
		if !user.CanAdmin() {
			return nil, ErrForbidden
		}
		if req, err = b.promptRequest(ctx, user, thread.Dir); err != nil {
			return nil, err
		}
	}
	if err := b.Notes.Edit(ctx, note.ID, body); err != nil {
		return nil, err
	}
	if note, err = b.Notes.Get(ctx, note.ID); err != nil {
		return nil, err
	}
	if note.ID == thread.ID {
		thread = note
	}
	b.publish(note)
	instruction := b.instructs(user, note.Body)
	if !thread.IsPrompt() && !instruction {
		return note, nil
	}
	if err := b.Notes.SetState(ctx, thread.ID, model.NoteOpen, ""); err != nil {
		return nil, err
	}
	thread.State = model.NoteOpen
	b.publish(thread)
	again := b.answerTo(ctx, thread, note)
	switch {
	case thread.IsPrompt():
		b.dispatchPrompt(thread, req, again)
	case b.AgentJob != "":
		b.dispatch(thread, again)
	}
	return note, nil
}

// answerTo returns the agent's answer to a note of a thread, the first
// reply of the agent after it, or nil for none yet.
func (b *Board) answerTo(ctx context.Context, thread, note *model.Note) *model.Note {
	replies, _ := b.Notes.Replies(ctx, []string{thread.ID})
	after := note.ID == thread.ID
	for _, reply := range replies[thread.ID] {
		if after && reply.UserName == b.Answerer() {
			return &reply
		}
		if reply.ID == note.ID {
			after = true
		}
	}
	return nil
}

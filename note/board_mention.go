package note

import (
	"context"

	"github.com/titpetric/task-ui/model"
)

// instructs reports whether a note of the user's with the body is an
// instruction: it names a user other than its author, the agent or one
// the board knows, and the author may run jobs, as a note from someone
// who may not is a message. A name of no user is a plain mention.
func (b *Board) instructs(ctx context.Context, user *model.User, body string) bool {
	named := false
	for _, name := range model.Mentions(body) {
		if name != user.Name && b.known(ctx, name) {
			named = true
			break
		}
	}
	return named && b.permissions(ctx, user).Allowed("run")
}

// known reports whether a name is the agent's or a user's.
func (b *Board) known(ctx context.Context, name string) bool {
	if b.Agent != "" && name == b.Agent {
		return true
	}
	if b.Users == nil {
		return false
	}
	_, err := b.Users.GetByName(ctx, name)
	return err == nil
}

// forAgent reports whether an instruction of the user's with the body is
// the board's to run: it names the agent, someone else wrote it, and
// there is an agent job.
func (b *Board) forAgent(user *model.User, body string) bool {
	return b.AgentJob != "" && b.Agent != "" && user.Name != b.Agent && contains(model.Mentions(body), b.Agent)
}

// threadForAgent reports whether an open thread is the agent's to run:
// its note or a reply by someone else names the agent.
func (b *Board) threadForAgent(ctx context.Context, note *model.Note) bool {
	if b.AgentJob == "" || b.Agent == "" {
		return false
	}
	if note.UserName != b.Agent && note.Mentioned(b.Agent) {
		return true
	}
	replies, _ := b.Notes.Replies(ctx, []string{note.ID})
	for _, reply := range replies[note.ID] {
		if reply.UserName != b.Agent && reply.Mentioned(b.Agent) {
			return true
		}
	}
	return false
}

// assigned reports whether the user may set an instruction's state: the
// note or a reply in its thread names the user, the user wrote the note,
// or the user is an admin.
func (b *Board) assigned(ctx context.Context, user *model.User, note *model.Note) bool {
	if user == nil {
		return false
	}
	if user.CanAdmin() || note.UserName == user.Name || note.Mentioned(user.Name) {
		return true
	}
	replies, _ := b.Notes.Replies(ctx, []string{note.ID})
	for _, reply := range replies[note.ID] {
		if reply.Mentioned(user.Name) {
			return true
		}
	}
	return false
}

// contains reports whether names holds name.
func contains(names []string, name string) bool {
	for _, n := range names {
		if n == name {
			return true
		}
	}
	return false
}

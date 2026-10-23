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

// known reports whether a name is a user's.
func (b *Board) known(ctx context.Context, name string) bool {
	if b.Users == nil {
		return false
	}
	_, err := b.Users.GetByName(ctx, name)
	return err == nil
}

// agentIn returns the first user a text by the author names, other
// than the author, whose mentions run the Notes agent, Run instance
// Claude when mentioned on, or nil.
func (b *Board) agentIn(ctx context.Context, author, body string) *model.User {
	if b.Users == nil {
		return nil
	}
	for _, name := range model.Mentions(body) {
		if name == author {
			continue
		}
		if user, err := b.Users.GetByName(ctx, name); err == nil && user.AgentReply && user.Mcp {
			return user
		}
	}
	return nil
}

// forAgent returns the agent an instruction of the user's with the body
// is the board's to run as, or nil: it names a user with Run instance
// Claude when mentioned, someone else wrote it. A note naming any other
// user waits for that user, their agent over MCP or themselves.
func (b *Board) forAgent(ctx context.Context, user *model.User, body string) *model.User {
	return b.agentIn(ctx, user.Name, body)
}

// threadForAgent returns the agent an open thread is for, or nil: its
// note or a reply by someone else names a user with Run instance Claude
// when mentioned.
func (b *Board) threadForAgent(ctx context.Context, note *model.Note) *model.User {
	if agent := b.agentIn(ctx, note.UserName, note.Body); agent != nil {
		return agent
	}
	replies, _ := b.Notes.Replies(ctx, []string{note.ID})
	for _, reply := range replies[note.ID] {
		if agent := b.agentIn(ctx, reply.UserName, reply.Body); agent != nil {
			return agent
		}
	}
	return nil
}

// AnswererOf returns the user name the agent's replies on a thread
// carry: claude on a prompt, the agent the thread names on an
// instruction, else empty.
func (b *Board) AnswererOf(ctx context.Context, thread *model.Note) string {
	if thread.IsPrompt() {
		return answerer
	}
	if agent := b.threadForAgent(ctx, thread); agent != nil {
		return agent.Name
	}
	return ""
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

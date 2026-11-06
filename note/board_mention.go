package note

import (
	"context"

	"github.com/titpetric/task-ui/model"
)

// instructs reports whether a note of the user's with the body is an
// instruction: it addresses a user the board knows, and its author may
// run jobs. The thread is the note a reply answers, nil for a new note.
func (b *Board) instructs(ctx context.Context, user *model.User, thread *model.Note, body string) bool {
	return len(b.addressed(ctx, user.Name, thread, body)) > 0 && b.permissions(ctx, user).Allowed("run")
}

// addressed returns the users a message of the author's addresses, by
// name: those its body names with @, the author apart, that the board
// knows. A reply naming nobody takes the names of its note, and a reply
// from a user the thread names takes none and reports instead.
func (b *Board) addressed(ctx context.Context, author string, thread *model.Note, body string) []string {
	if names := b.named(ctx, author, body); len(names) > 0 || thread == nil {
		return names
	}
	if b.addresses(ctx, thread, author) {
		return nil
	}
	return b.named(ctx, author, thread.Body)
}

// addresses reports whether a thread names a user, in its note or in one
// of its replies.
func (b *Board) addresses(ctx context.Context, thread *model.Note, name string) bool {
	if thread.Mentioned(name) {
		return true
	}
	replies, _ := b.Notes.Replies(ctx, []string{thread.ID})
	for _, reply := range replies[thread.ID] {
		if reply.Mentioned(name) {
			return true
		}
	}
	return false
}

// named returns the users a text names with @, the author apart, that
// the board knows, once each in order.
func (b *Board) named(ctx context.Context, author, text string) []string {
	var names []string
	for _, name := range model.Mentions(text) {
		if name != author && b.known(ctx, name) {
			names = append(names, name)
		}
	}
	return names
}

// known reports whether a name is a user's.
func (b *Board) known(ctx context.Context, name string) bool {
	if b.Users == nil {
		return false
	}
	_, err := b.Users.GetByName(ctx, name)
	return err == nil
}

// agents returns the users of the names whose mentions run the Notes
// agent, Run instance Claude when mentioned on.
func (b *Board) agents(ctx context.Context, names []string) []*model.User {
	if b.Users == nil {
		return nil
	}
	var agents []*model.User
	for _, name := range names {
		if user, err := b.Users.GetByName(ctx, name); err == nil && user.AgentReply && user.Mcp {
			agents = append(agents, user)
		}
	}
	return agents
}

// forAgents returns the agents an instruction of the user's with the
// body is for, a run each. A message naming any other user waits for
// that user, their agent over MCP, or themselves.
func (b *Board) forAgents(ctx context.Context, user *model.User, thread *model.Note, body string) []*model.User {
	return b.agents(ctx, b.addressed(ctx, user.Name, thread, body))
}

// threadAgents returns the agents an open thread is for, once each: the
// ones its note names and the ones its replies name.
func (b *Board) threadAgents(ctx context.Context, note *model.Note) []*model.User {
	var agents []*model.User
	var names []string
	add := func(author, body string) {
		for _, agent := range b.agents(ctx, b.named(ctx, author, body)) {
			if !contains(names, agent.Name) {
				names = append(names, agent.Name)
				agents = append(agents, agent)
			}
		}
	}
	add(note.UserName, note.Body)
	replies, _ := b.Notes.Replies(ctx, []string{note.ID})
	for _, reply := range replies[note.ID] {
		add(reply.UserName, reply.Body)
	}
	return agents
}

// threadForAgent returns the first agent a thread names, or nil. It
// stops there, as a page asks it of every thread it lists.
func (b *Board) threadForAgent(ctx context.Context, note *model.Note) *model.User {
	if agents := b.agents(ctx, b.named(ctx, note.UserName, note.Body)); len(agents) > 0 {
		return agents[0]
	}
	replies, _ := b.Notes.Replies(ctx, []string{note.ID})
	for _, reply := range replies[note.ID] {
		if agents := b.agents(ctx, b.named(ctx, reply.UserName, reply.Body)); len(agents) > 0 {
			return agents[0]
		}
	}
	return nil
}

// AnswererOf returns the user name the agent's replies on a thread are
// from: claude on a prompt, the agent the thread names on an
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
// thread names the user, the user wrote the note, or the user is an
// admin.
func (b *Board) assigned(ctx context.Context, user *model.User, note *model.Note) bool {
	if user == nil {
		return false
	}
	return user.CanAdmin() || note.UserName == user.Name || b.addresses(ctx, note, user.Name)
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

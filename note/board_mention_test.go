package note_test

import (
	"context"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
)

// users adds runners to the fixture's users, by name.
func (f *fixture) users(t *testing.T, names ...string) []*model.User {
	t.Helper()
	list := make([]*model.User, 0, len(names))
	for _, name := range names {
		user := &model.User{Name: name, Role: model.RoleRunner}
		assert.NoError(t, f.board.Users.Create(context.Background(), user))
		list = append(list, user)
	}
	return list
}

// A note naming a user other than the agent is an instruction to that
// user: it stays open for the user to take over MCP, the board runs
// nothing, and only the user it names, its author or an admin sets its
// state. A viewer's note to the user is a message.
func TestBoard_InstructionToUser(t *testing.T) {
	t.Parallel()
	f := newFixture(t, "agent/run")
	ctx := context.Background()
	named := f.users(t, "consuela", "rob")
	consuela, rob := named[0], named[1]

	open, err := f.board.Post(ctx, f.ops, "", "@consuela tidy the README")
	assert.NoError(t, err)
	assert.Equal(t, model.NoteOpen, open.State)
	time.Sleep(100 * time.Millisecond)
	got, _ := f.board.Notes.Get(ctx, open.ID)
	assert.Equal(t, model.NoteOpen, got.State, "the board runs nothing for consuela")
	assert.Equal(t, "", got.RunID)
	assert.Len(t, replies(t, f, open.ID), 0)
	mine, err := f.board.Notes.List(ctx, storage.NoteFilter{Channel: model.ChannelNotes, Mention: "consuela", State: model.NoteOpen})
	assert.NoError(t, err)
	assert.Len(t, mine, 1)

	_, err = f.board.Comment(ctx, rob, open.ID, "", model.NoteDone)
	assert.ErrorIs(t, err, note.ErrNotAssigned)
	_, err = f.board.Comment(ctx, rob, open.ID, "I can help.", "")
	assert.NoError(t, err, "anyone may reply")
	_, err = f.board.Comment(ctx, consuela, open.ID, "", model.NoteWorking)
	assert.NoError(t, err)
	_, err = f.board.Comment(ctx, consuela, open.ID, "Tidied in abc123.", model.NoteDone)
	assert.NoError(t, err)
	got, _ = f.board.Notes.Get(ctx, open.ID)
	assert.Equal(t, model.NoteDone, got.State)
	_, err = f.board.Comment(ctx, f.ops, open.ID, "", model.NoteFailed)
	assert.NoError(t, err, "the author sets the state too")

	viewer := &model.User{Name: "guest", Role: model.RoleViewer}
	assert.NoError(t, f.board.Users.Create(ctx, viewer))
	watched, err := f.board.Post(ctx, viewer, "", "@consuela delete everything")
	assert.NoError(t, err)
	assert.Equal(t, "", watched.State, "a viewer cannot hand out work")
	_, err = f.board.Comment(ctx, consuela, watched.ID, "", model.NoteDone)
	assert.ErrorIs(t, err, note.ErrInvalid, "not an instruction")
}

// A reply naming a user opens a done note again for that user, without
// the agent, though the thread started with it.
func TestBoard_ReplyToUser(t *testing.T) {
	t.Parallel()
	f := newFixture(t, "agent/run")
	ctx := context.Background()
	f.users(t, "consuela")
	first, err := f.board.Post(ctx, f.ops, "", "@claude update the weather job")
	assert.NoError(t, err)
	assert.Equal(t, model.NoteDone, f.settled(t, first.ID).State)

	_, err = f.board.Post(ctx, f.ops, first.ID, "@consuela check what it did")
	assert.NoError(t, err)
	got, _ := f.board.Notes.Get(ctx, first.ID)
	assert.Equal(t, model.NoteOpen, got.State)
	time.Sleep(100 * time.Millisecond)
	assert.Len(t, replies(t, f, first.ID), 2, "the agent does not answer")
	mine, err := f.board.Notes.List(ctx, storage.NoteFilter{Channel: model.ChannelNotes, Mention: "consuela", State: model.NoteOpen})
	assert.NoError(t, err)
	assert.Len(t, mine, 1, "the reply names consuela")
}

// After a restart, a note an external agent set working, with no run,
// goes on, and an open note to that user starts nothing.
func TestBoard_RecoverExternal(t *testing.T) {
	t.Parallel()
	f := newFixture(t, "agent/run")
	ctx := context.Background()
	consuela := f.users(t, "consuela")[0]
	working, err := f.board.Post(ctx, f.ops, "", "@consuela go")
	assert.NoError(t, err)
	_, err = f.board.Comment(ctx, consuela, working.ID, "", model.NoteWorking)
	assert.NoError(t, err)
	open, err := f.board.Post(ctx, f.ops, "", "@consuela and this")
	assert.NoError(t, err)

	assert.NoError(t, f.board.Recover(ctx))
	time.Sleep(100 * time.Millisecond)
	got, _ := f.board.Notes.Get(ctx, working.ID)
	assert.Equal(t, model.NoteWorking, got.State)
	assert.Len(t, replies(t, f, working.ID), 0)
	got, _ = f.board.Notes.Get(ctx, open.ID)
	assert.Equal(t, model.NoteOpen, got.State)
	assert.Equal(t, "", got.RunID)
}

// A thread the agent answered keeps its run; a reply that names
// another user, who sets it working, lets go of that run, so the note
// stays working for that user across a restart.
func TestBoard_RecoverHandedOver(t *testing.T) {
	t.Parallel()
	f := newFixture(t, "agent/run")
	ctx := context.Background()
	consuela := f.users(t, "consuela")[0]
	first, err := f.board.Post(ctx, f.ops, "", "@claude update the weather job")
	assert.NoError(t, err)
	assert.NotEqual(t, "", f.settled(t, first.ID).RunID)

	_, err = f.board.Post(ctx, f.ops, first.ID, "@consuela check what it did")
	assert.NoError(t, err)
	_, err = f.board.Comment(ctx, consuela, first.ID, "", model.NoteWorking)
	assert.NoError(t, err)
	got, _ := f.board.Notes.Get(ctx, first.ID)
	assert.Equal(t, "", got.RunID, "no run of the board's works on it")

	restarted := *f.board
	restarted.Runner = runner.New(runner.Options{Runs: f.board.Runs, Grace: time.Second, KillTimeout: time.Second, Keep: time.Minute})
	assert.NoError(t, restarted.Recover(ctx))
	time.Sleep(100 * time.Millisecond)
	got, _ = f.board.Notes.Get(ctx, first.ID)
	assert.Equal(t, model.NoteWorking, got.State)
	assert.Len(t, replies(t, f, first.ID), 2, "nothing fails the note")
}

//go:build integration

package note_test

import (
	"context"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/event"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
	"github.com/titpetric/task-ui/storage"
)

// noRoot is what a dispatch leaves on a note when the board has no root
// to run the Notes agent in, and a test's sign that the board dispatched.
const noRoot = "has no root to run in"

// newBoard is a board over its own in-memory database, with the users
// ops, an admin who may run jobs, the agents claude and consuela, and
// alice, who is neither. The board has no root, so every dispatch fails
// at once and leaves a reply as the agent it dispatched: that reply is
// what a test counts. Its event bus holds every state the board wrote,
// which watch reads back.
func newBoard(t *testing.T) (*note.Board, *storage.NoteStorage, map[string]*model.User) {
	t.Helper()
	ctx := context.Background()
	db, err := storage.OpenAndMigrate(ctx, ":memory:", app.Migrations())
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	notes, users := storage.NewNoteStorage(db), storage.NewUserStorage(db)
	people := map[string]*model.User{
		"ops":      {Name: "ops", Role: model.RoleAdmin},
		"claude":   {Name: "claude", Role: model.RoleRunner, Mcp: true, AgentReply: true},
		"consuela": {Name: "consuela", Role: model.RoleRunner, Mcp: true, AgentReply: true},
		"alice":    {Name: "alice", Role: model.RoleRunner},
	}
	for _, user := range people {
		assert.NoError(t, users.Create(ctx, user))
	}
	return &note.Board{Notes: notes, Users: users, Events: event.NewBus()}, notes, people
}

// watch subscribes to the board's notes. It returns the states one note
// went through, in the order the board published them.
func watch(board *note.Board) func(id string) []string {
	events, cancel := board.Events.Subscribe(note.Topic)
	return func(id string) []string {
		cancel()
		var out []string
		for e := range events {
			if row, ok := e.Data.(model.Note); ok && row.ID == id {
				out = append(out, row.State)
			}
		}
		return out
	}
}

// answers returns the user names of the replies to a note, oldest first.
func answers(t *testing.T, notes *storage.NoteStorage, id string) []string {
	t.Helper()
	replies, err := notes.Replies(context.Background(), []string{id})
	assert.NoError(t, err)
	var names []string
	for _, reply := range replies[id] {
		names = append(names, reply.UserName)
	}
	return names
}

// A reply in a thread addresses the users its note names, so the agent
// runs again without being named a second time; a reply that names
// someone addresses that user alone, and a reply from a user the note
// names reports on the thread and starts nothing.
func TestBoard_Post(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	board, notes, people := newBoard(t)

	thread, err := board.Post(ctx, people["ops"], "", "@claude update the weather job")
	assert.NoError(t, err)
	assert.Equal(t, model.NoteOpen, thread.State)
	assert.Equal(t, []string{"claude"}, answers(t, notes, thread.ID))

	_, err = board.Post(ctx, people["ops"], thread.ID, "and add Lisbon")
	assert.NoError(t, err)
	assert.Equal(t, []string{"claude", "ops", "claude"}, answers(t, notes, thread.ID), "the reply carries the note's name")

	_, err = board.Post(ctx, people["ops"], thread.ID, "@consuela take this one")
	assert.NoError(t, err)
	assert.Equal(t, []string{"claude", "ops", "claude", "ops", "consuela"}, answers(t, notes, thread.ID), "a named user is the only one addressed")

	_, err = board.Post(ctx, people["claude"], thread.ID, "the job is updated")
	assert.NoError(t, err)
	assert.Equal(t, []string{"claude", "ops", "claude", "ops", "consuela", "claude"}, answers(t, notes, thread.ID), "a report starts nothing")

	_, err = board.Post(ctx, people["consuela"], thread.ID, "and the archive is smaller")
	assert.NoError(t, err)
	assert.Equal(t, []string{"claude", "ops", "claude", "ops", "consuela", "claude", "consuela"}, answers(t, notes, thread.ID), "a user a reply names reports too")

	plain, err := board.Post(ctx, people["ops"], "", "the weather job is slow")
	assert.NoError(t, err)
	assert.Equal(t, "", plain.State)
	_, err = board.Post(ctx, people["ops"], plain.ID, "it reads the whole archive")
	assert.NoError(t, err)
	assert.Equal(t, []string{"ops"}, answers(t, notes, plain.ID), "a thread that names nobody stays a message")
}

// A message that names several agents hands each of them the work, and
// each gets a run and a reply of its own; a reply of the thread's does
// the same.
func TestBoard_Comment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	board, notes, people := newBoard(t)

	thread, err := board.Post(ctx, people["ops"], "", "@claude take the API, @consuela take the docs")
	assert.NoError(t, err)
	assert.Equal(t, []string{"claude", "consuela"}, answers(t, notes, thread.ID))
	replies, err := notes.Replies(ctx, []string{thread.ID})
	assert.NoError(t, err)
	for _, reply := range replies[thread.ID] {
		assert.Contains(t, reply.Body, noRoot, reply.UserName)
	}

	_, err = board.Comment(ctx, people["consuela"], thread.ID, "the docs are done", model.NoteWorking)
	assert.NoError(t, err)
	assert.Equal(t, []string{"claude", "consuela", "consuela"}, answers(t, notes, thread.ID), "an agent's report addresses nobody")

	_, err = board.Comment(ctx, people["ops"], thread.ID, "both of you, once more", "")
	assert.NoError(t, err)
	assert.Equal(t, []string{"claude", "consuela", "consuela", "ops", "claude", "consuela"}, answers(t, notes, thread.ID), "the reply carries both names")
}

// The answerer of a thread is the first agent it names, claude on a
// prompt, and nobody on a thread that names nobody.
func TestBoard_AnswererOf(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	board, notes, people := newBoard(t)

	thread := &model.Note{UserName: "ops", Body: "@alice and @consuela, have a look", State: model.NoteOpen}
	assert.NoError(t, notes.Create(ctx, thread))
	assert.Equal(t, "consuela", board.AnswererOf(ctx, thread))

	message := &model.Note{UserName: "ops", Body: "the archive is big"}
	assert.NoError(t, notes.Create(ctx, message))
	assert.Equal(t, "", board.AnswererOf(ctx, message))
	assert.NoError(t, notes.Create(ctx, &model.Note{ParentID: message.ID, UserName: "ops", Body: "@claude look at it"}))
	assert.Equal(t, "claude", board.AnswererOf(ctx, message), "a reply names the agent")

	assert.Equal(t, board.Answerer(), board.AnswererOf(ctx, &model.Note{WorkspaceID: model.Default, Body: "say ok"}))
	assert.NotNil(t, people["alice"])
}

// Every agent a thread names answers it, once each: the ones its note
// names and the ones its replies name. A prompt is claude's alone.
func TestBoard_AnswerersOf(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	board, notes, _ := newBoard(t)

	thread := &model.Note{UserName: "ops", Body: "@claude take the API, @consuela take the docs", State: model.NoteOpen}
	assert.NoError(t, notes.Create(ctx, thread))
	assert.Equal(t, []string{"claude", "consuela"}, board.AnswerersOf(ctx, thread))

	assert.NoError(t, notes.Create(ctx, &model.Note{ParentID: thread.ID, UserName: "ops", Body: "@claude again"}))
	assert.Equal(t, []string{"claude", "consuela"}, board.AnswerersOf(ctx, thread), "a name twice answers once")

	message := &model.Note{UserName: "ops", Body: "@alice, have a look"}
	assert.NoError(t, notes.Create(ctx, message))
	assert.Len(t, board.AnswerersOf(ctx, message), 0, "alice runs no agent")
	assert.NoError(t, notes.Create(ctx, &model.Note{ParentID: message.ID, UserName: "ops", Body: "@consuela too"}))
	assert.Equal(t, []string{"consuela"}, board.AnswerersOf(ctx, message), "a reply names the agent")

	assert.Equal(t, []string{board.Answerer()}, board.AnswerersOf(ctx, &model.Note{WorkspaceID: model.Default, Body: "say ok"}))
}

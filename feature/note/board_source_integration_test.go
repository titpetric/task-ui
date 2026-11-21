//go:build integration

package note_test

import (
	"context"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/event"
	"github.com/titpetric/task-ui/feature/note"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/storage"
)

// A run's reply fills the answer its thread holds, in place of a new reply.
func TestBoard_AnswerThread(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := storage.OpenAndMigrate(ctx, ":memory:", app.Migrations())
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	notes, users := storage.NewNoteStorage(db), storage.NewUserStorage(db)
	board := &note.Board{Notes: notes, Users: users, Events: event.NewBus()}
	agent := &model.User{Name: "claude", Role: model.RoleRunner, AgentReply: true, Mcp: true}
	person := &model.User{Name: "ops", Role: model.RoleAdmin}
	assert.NoError(t, users.Create(ctx, agent))
	assert.NoError(t, users.Create(ctx, person))

	thread := &model.Note{UserName: "ops", Body: "why is the build red", State: model.NoteWorking, WorkspaceID: model.Default, Dir: "infra"}
	assert.NoError(t, notes.Create(ctx, thread))
	answer := &model.Note{ParentID: thread.ID, UserName: "claude", Agent: "claude", State: model.NoteWorking, RunID: model.NewID()}
	assert.NoError(t, notes.CreateAnswer(ctx, answer))

	reply, err := board.AnswerThread(ctx, agent, thread.ID, "because the cache was cold", "")
	assert.NoError(t, err)
	assert.Equal(t, answer.ID, reply.ID, "the answer the run holds is filled")
	assert.Equal(t, "because the cache was cold", reply.Body)
}

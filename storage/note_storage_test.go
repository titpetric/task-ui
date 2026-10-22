package storage_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// A note keeps its mentions; the board lists threads newest first, by
// channel, mention and state, and the replies come with them.
func TestNoteStorage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	notes := storage.NewNoteStorage(newDB(t))

	assert.Error(t, notes.Create(ctx, &model.Note{UserName: "ops", Body: "  "}))
	task := &model.Note{UserName: "ops", Body: "@claude update the weather job", State: model.NoteOpen}
	assert.NoError(t, notes.Create(ctx, task))
	assert.Equal(t, model.ChannelNotes, task.Channel)
	assert.True(t, task.Mentioned("claude"))
	other := &model.Note{UserName: "ops", Body: "@alice have a look"}
	assert.NoError(t, notes.Create(ctx, other))
	chat := &model.Note{Channel: "chat", UserName: "alice", Body: "hello"}
	assert.NoError(t, notes.Create(ctx, chat))
	reply := &model.Note{UserName: "claude", Body: "done", ParentID: task.ID}
	assert.NoError(t, notes.Create(ctx, reply))

	board, err := notes.List(ctx, storage.NoteFilter{Channel: model.ChannelNotes})
	assert.NoError(t, err)
	assert.Len(t, board, 2)
	assert.Equal(t, other.ID, board[0].ID, "newest first")

	mine, _ := notes.List(ctx, storage.NoteFilter{Mention: "claude", State: model.NoteOpen})
	assert.Len(t, mine, 1)
	assert.Equal(t, task.ID, mine[0].ID)

	replies, err := notes.Replies(ctx, []string{task.ID, other.ID})
	assert.NoError(t, err)
	assert.Len(t, replies[task.ID], 1)
	assert.Len(t, replies[other.ID], 0)
	empty, _ := notes.Replies(ctx, nil)
	assert.Len(t, empty, 0)

	assert.NoError(t, notes.SetState(ctx, task.ID, model.NoteWorking, "run1"))
	assert.NoError(t, notes.SetState(ctx, task.ID, model.NoteDone, ""))
	got, err := notes.Get(ctx, task.ID)
	assert.NoError(t, err)
	assert.Equal(t, model.NoteDone, got.State)
	assert.Equal(t, "run1", got.RunID, "kept")
	assert.ErrorIs(t, notes.SetState(ctx, "missing", model.NoteDone, ""), storage.ErrNotFound)

	assert.NoError(t, notes.Delete(ctx, task.ID))
	_, err = notes.Get(ctx, reply.ID)
	assert.ErrorIs(t, err, storage.ErrNotFound, "the replies go too")
}

// A prompt is a note with a workspace: the board's listing leaves it out,
// and the prompts' listing narrows by workspace and place.
func TestNoteStorage_Prompts(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	notes := storage.NewNoteStorage(newDB(t))
	assert.NoError(t, notes.Create(ctx, &model.Note{UserName: "ops", Body: "a note"}))
	here := &model.Note{UserName: "admin", Body: "what builds here", State: model.NoteOpen, WorkspaceID: "ws1", Dir: "malta"}
	assert.NoError(t, notes.Create(ctx, here))
	assert.NoError(t, notes.Create(ctx, &model.Note{UserName: "admin", Body: "and here", State: model.NoteOpen, WorkspaceID: "ws1", Dir: "malta/dockerfiles"}))

	board, _ := notes.List(ctx, storage.NoteFilter{State: model.NoteOpen})
	assert.Len(t, board, 0)
	all, _ := notes.List(ctx, storage.NoteFilter{Prompts: true})
	assert.Len(t, all, 2)
	mine, _ := notes.List(ctx, storage.NoteFilter{Prompts: true, WorkspaceID: "ws1", Dir: "malta"})
	assert.Len(t, mine, 1)
	assert.Equal(t, here.ID, mine[0].ID)
	none, _ := notes.List(ctx, storage.NoteFilter{Prompts: true, WorkspaceID: "ws2"})
	assert.Len(t, none, 0)
}

//go:build integration

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

	byAlice, _ := notes.List(ctx, storage.NoteFilter{UserName: "alice"})
	assert.Len(t, byAlice, 1)
	assert.Equal(t, chat.ID, byAlice[0].ID)

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
	assert.NoError(t, notes.Release(ctx, task.ID, model.NoteWorking))
	got, _ = notes.Get(ctx, task.ID)
	assert.Equal(t, model.NoteWorking, got.State)
	assert.Equal(t, "", got.RunID, "let go")

	assert.NoError(t, notes.Delete(ctx, task.ID))
	_, err = notes.Get(ctx, reply.ID)
	assert.ErrorIs(t, err, storage.ErrNotFound, "the replies go too")
}

// A prompt is a note with a workspace: the board's listing leaves it out,
// and the prompts' listing narrows by workspace and place, or by the
// run a prompt asks about.
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

	// A prompt about a run lists by the run, not with its place's.
	about := &model.Note{UserName: "admin", Body: "why did it fail", State: model.NoteOpen, WorkspaceID: "ws1", Dir: "malta", AboutRunID: "run1"}
	assert.NoError(t, notes.Create(ctx, about))
	mine, _ = notes.List(ctx, storage.NoteFilter{Prompts: true, Dir: "malta"})
	assert.Len(t, mine, 1)
	assert.Equal(t, here.ID, mine[0].ID)
	run, _ := notes.List(ctx, storage.NoteFilter{Prompts: true, AboutRunID: "run1"})
	assert.Len(t, run, 1)
	assert.Equal(t, about.ID, run[0].ID)
	all, _ = notes.List(ctx, storage.NoteFilter{Prompts: true})
	assert.Len(t, all, 3)

	// A workflow run's prompt lists by the workflow run, not with its
	// place's.
	flow := &model.Note{UserName: "admin", Body: "summarize", State: model.NoteOpen, WorkspaceID: "ws1", Dir: "malta", AboutWorkflowRunID: "wr1"}
	assert.NoError(t, notes.Create(ctx, flow))
	mine, _ = notes.List(ctx, storage.NoteFilter{Prompts: true, Dir: "malta"})
	assert.Len(t, mine, 1)
	assert.Equal(t, here.ID, mine[0].ID)
	wr, _ := notes.List(ctx, storage.NoteFilter{Prompts: true, AboutWorkflowRunID: "wr1"})
	assert.Len(t, wr, 1)
	assert.Equal(t, flow.ID, wr[0].ID)
	assert.Equal(t, "wr1", wr[0].AboutWorkflowRunID)
}

// Edit changes a note's body and mentions and marks it edited; Answer
// rewrites a reply in place, with its state and run, unmarked.
func TestNoteStorage_Edit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	notes := storage.NewNoteStorage(newDB(t))
	posted := &model.Note{UserName: "ops", Body: "have a look"}
	assert.NoError(t, notes.Create(ctx, posted))
	assert.Nil(t, posted.EditedAt)

	assert.NoError(t, notes.Edit(ctx, posted.ID, "  @alice have a look  "))
	got, err := notes.Get(ctx, posted.ID)
	assert.NoError(t, err)
	assert.Equal(t, "@alice have a look", got.Body)
	assert.True(t, got.Mentioned("alice"))
	assert.NotNil(t, got.EditedAt)
	assert.Error(t, notes.Edit(ctx, posted.ID, " "))
	assert.ErrorIs(t, notes.Edit(ctx, "missing", "x"), storage.ErrNotFound)

	reply := &model.Note{UserName: "claude", Body: "first", ParentID: posted.ID, RunID: "r1"}
	assert.NoError(t, notes.Create(ctx, reply))
	assert.NoError(t, notes.Answer(ctx, reply.ID, "first", model.NoteWorking, ""))
	got, _ = notes.Get(ctx, reply.ID)
	assert.Equal(t, model.NoteWorking, got.State)
	assert.Equal(t, "r1", got.RunID)
	assert.NoError(t, notes.Answer(ctx, reply.ID, "second", "", "r2"))
	got, _ = notes.Get(ctx, reply.ID)
	assert.Equal(t, "second", got.Body)
	assert.Equal(t, "", got.State)
	assert.Equal(t, "r2", got.RunID)
	assert.Nil(t, got.EditedAt, "an answer run again is no edit")
	assert.ErrorIs(t, notes.Answer(ctx, "missing", "x", "", ""), storage.ErrNotFound)
}

// A filter by area lists an area's prompts; a filter by place leaves
// them out.
func TestNoteStorage_Area(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	notes := storage.NewNoteStorage(newDB(t))
	assert.NoError(t, notes.Create(ctx, &model.Note{UserName: "ops", Body: "build one", WorkspaceID: "_", Dir: "_", Area: model.AreaWorkflows}))
	assert.NoError(t, notes.Create(ctx, &model.Note{UserName: "ops", Body: "a root prompt", WorkspaceID: "_", Dir: "_"}))
	area, err := notes.List(ctx, storage.NoteFilter{Prompts: true, Area: model.AreaWorkflows})
	assert.NoError(t, err)
	assert.Len(t, area, 1)
	assert.Equal(t, "build one", area[0].Body)
	root, _ := notes.List(ctx, storage.NoteFilter{Prompts: true, Dir: "_"})
	assert.Len(t, root, 1)
	assert.Equal(t, "a root prompt", root[0].Body)
}

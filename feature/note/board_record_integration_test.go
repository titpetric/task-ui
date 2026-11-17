//go:build integration

package note_test

import (
	"context"
	"errors"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/feature/note"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/storage"
)

// A spoken note holds its words and its clip, starts no run, and a clip with
// no words still leaves a message to play.
func TestBoard_RecordSaves(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	board, runs, ops := askingBoard(t)

	recorded, err := board.Record(ctx, ops, "", "ship it", "voice/2026-11-16T10-00-00.wav")
	assert.NoError(t, err)
	assert.Equal(t, model.ChannelNotes, recorded.Channel)
	assert.Equal(t, "ship it", recorded.Body)
	assert.Equal(t, "ship it", recorded.Spoken)
	assert.Equal(t, "voice/2026-11-16T10-00-00.wav", recorded.AudioPath)
	assert.Equal(t, "", recorded.WorkspaceID)
	assert.Equal(t, "", recorded.State)
	assert.Equal(t, "", recorded.ParentID)

	held, err := board.Notes.Get(ctx, recorded.ID)
	assert.NoError(t, err)
	assert.Equal(t, recorded.AudioPath, held.AudioPath)

	empty, err := board.Record(ctx, ops, "", "", "voice/silent.wav")
	assert.NoError(t, err)
	assert.Equal(t, "A recording.", empty.Body)
	assert.Equal(t, "", empty.Spoken)

	_, err = board.Record(ctx, ops, "", "", "")
	assert.ErrorIs(t, err, note.ErrInvalid)

	runsList, err := runs.List(ctx, storage.RunFilter{})
	assert.NoError(t, err)
	assert.Len(t, runsList, 0, "a spoken note starts no run")
}

// A later clip replies in the note thread; a reply to a reply and a prompt
// are refused.
func TestBoard_RecordReply(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	board, _, ops := askingBoard(t)

	root, err := board.Record(ctx, ops, "", "first", "voice/first.wav")
	assert.NoError(t, err)
	reply, err := board.Record(ctx, ops, root.ID, "second", "voice/second.wav")
	assert.NoError(t, err)
	assert.Equal(t, root.ID, reply.ParentID)
	assert.Equal(t, model.ChannelNotes, reply.Channel)

	replies, err := board.Notes.Replies(ctx, []string{root.ID})
	assert.NoError(t, err)
	assert.Len(t, replies[root.ID], 1)

	_, err = board.Record(ctx, ops, reply.ID, "third", "")
	assert.ErrorIs(t, err, note.ErrInvalid)

	prompt := &model.Note{Channel: model.ChannelNotes, UserID: ops.ID, UserName: ops.Name, Body: "a prompt", WorkspaceID: "infra"}
	assert.NoError(t, board.Notes.Create(ctx, prompt))
	_, err = board.Record(ctx, ops, prompt.ID, "an answer", "")
	assert.ErrorIs(t, err, note.ErrInvalid)

	_, err = board.Record(ctx, ops, "no-such-note", "an answer", "")
	assert.ErrorIs(t, err, storage.ErrNotFound)
}

// A user without note/edit leaves no spoken note.
func TestBoard_RecordPermission(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	board, _, _ := askingBoard(t)

	_, err := board.Record(ctx, &model.User{Name: "guest"}, "", "hello", "")
	assert.ErrorIs(t, err, note.ErrNoteForbidden)
}

// An unknown parent is not found; the error is the storage's.
func TestBoard_RecordUnknown(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	board, _, ops := askingBoard(t)

	_, err := board.Record(ctx, ops, "no-such-note", "hello", "")
	if !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("want %v, got %v", storage.ErrNotFound, err)
	}
}

//go:build integration

package storage_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/storage"
)

// An event is one item of a run's work, keyed by the run and its sequence.
func TestNoteEventStorage_Append(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	events := storage.NewNoteEventStorage(newDB(t))
	event := &model.NoteEvent{NoteID: "n1", RunID: "r1", QuestionID: "q1", Seq: 1, Kind: "text", Body: "hello", Data: "{}"}

	assert.NoError(t, events.Append(ctx, event))
	assert.NotEqual(t, "", event.ID)
	assert.NoError(t, events.Append(ctx, &model.NoteEvent{NoteID: "n1", RunID: "r1", QuestionID: "q1", Seq: 1, Kind: "text", Body: "again"}))
}

// A thread's events are ordered by the message that asked, then the run and its sequence.
func TestNoteEventStorage_List(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	events := storage.NewNoteEventStorage(newDB(t))

	for _, event := range []*model.NoteEvent{
		{NoteID: "n1", RunID: "r1", QuestionID: "q1", Seq: 2, Kind: "text", Body: "second"},
		{NoteID: "n1", RunID: "r1", QuestionID: "q1", Seq: 1, Kind: "text", Body: "first"},
		{NoteID: "n2", RunID: "r2", QuestionID: "q2", Seq: 1, Kind: "text", Body: "other"},
	} {
		assert.NoError(t, events.Append(ctx, event))
	}

	got, err := events.List(ctx, "n1")
	assert.NoError(t, err)
	assert.Len(t, got, 2)
	assert.Equal(t, "first", got[0].Body)
	assert.Equal(t, "second", got[1].Body)
}

// A run's events are the tail a live page reads, oldest first.
func TestNoteEventStorage_ListRun(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	events := storage.NewNoteEventStorage(newDB(t))

	assert.NoError(t, events.Append(ctx, &model.NoteEvent{NoteID: "n1", RunID: "r1", Seq: 2, Kind: "text", Body: "second"}))
	assert.NoError(t, events.Append(ctx, &model.NoteEvent{NoteID: "n1", RunID: "r1", Seq: 1, Kind: "text", Body: "first"}))

	got, err := events.ListRun(ctx, "r1")
	assert.NoError(t, err)
	assert.Len(t, got, 2)
	assert.Equal(t, "first", got[0].Body)
}

// Last answers a run's state after a reload, or ErrNotFound.
func TestNoteEventStorage_Last(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	events := storage.NewNoteEventStorage(newDB(t))

	_, err := events.Last(ctx, "r1")
	assert.ErrorIs(t, err, storage.ErrNotFound)
	assert.NoError(t, events.Append(ctx, &model.NoteEvent{NoteID: "n1", RunID: "r1", Seq: 1, Kind: "text", Body: "first"}))
	assert.NoError(t, events.Append(ctx, &model.NoteEvent{NoteID: "n1", RunID: "r1", Seq: 2, Kind: "result", Body: "done"}))

	last, err := events.Last(ctx, "r1")
	assert.NoError(t, err)
	assert.Equal(t, "result", last.Kind)
}

// A rerun replaces a run's events, and a deleted thread takes its events with it.
func TestNoteEventStorage_DeleteByRun(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	events := storage.NewNoteEventStorage(newDB(t))

	assert.NoError(t, events.Append(ctx, &model.NoteEvent{NoteID: "n1", RunID: "r1", Seq: 1, Kind: "text", Body: "old"}))
	assert.NoError(t, events.Append(ctx, &model.NoteEvent{NoteID: "n1", RunID: "r2", Seq: 1, Kind: "text", Body: "other"}))
	assert.NoError(t, events.DeleteByRun(ctx, "r1"))

	got, err := events.List(ctx, "n1")
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, "r2", got[0].RunID)
}

// DeleteByNote frees every event of a thread.
func TestNoteEventStorage_DeleteByNote(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	events := storage.NewNoteEventStorage(newDB(t))

	assert.NoError(t, events.Append(ctx, &model.NoteEvent{NoteID: "n1", RunID: "r1", Seq: 1, Kind: "text"}))
	assert.NoError(t, events.DeleteByNote(ctx, "n1"))

	got, err := events.List(ctx, "n1")
	assert.NoError(t, err)
	assert.Len(t, got, 0)
}

// Every method returns the driver's error on a closed database.
func TestNoteEventStorage_Closed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	events := storage.NewNoteEventStorage(closedDB(t))

	assert.ErrorContains(t, events.Append(ctx, &model.NoteEvent{NoteID: "n1", RunID: "r1", Seq: 1}), errClosed)
	_, err := events.List(ctx, "n1")
	assert.ErrorContains(t, err, errClosed)
	_, err = events.ListRun(ctx, "r1")
	assert.ErrorContains(t, err, errClosed)
	_, err = events.Last(ctx, "r1")
	assert.ErrorContains(t, err, errClosed)
	assert.ErrorContains(t, events.DeleteByRun(ctx, "r1"), errClosed)
	assert.ErrorContains(t, events.DeleteByNote(ctx, "n1"), errClosed)
}

//go:build integration

package storage_test

import (
	"context"
	"testing"
	"time"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/storage"
)

// A meeting keeps its state, the participants who come and go, and the log they leave.
func TestMeetingStorage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	meetings := storage.NewMeetingStorage(newDB(t))
	now := time.Now().UTC().Truncate(time.Millisecond)

	room := &model.Meeting{ID: model.NewID(), Title: "Standup", UserID: "u1", UserName: "ana", State: model.MeetingOpen, StartedAt: &now, LimitSeconds: model.MeetingLimit}
	assert.NoError(t, meetings.Create(ctx, room))
	assert.True(t, room.CreatedAt != nil)

	got, err := meetings.Get(ctx, room.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Standup", got.Title)

	list, err := meetings.List(ctx, 10)
	assert.NoError(t, err)
	assert.Len(t, list, 1)

	open, err := meetings.Open(ctx)
	assert.NoError(t, err)
	assert.Len(t, open, 1)

	ana := &model.MeetingParticipant{MeetingID: room.ID, UserID: "u1", UserName: "ana"}
	assert.NoError(t, meetings.Join(ctx, ana))
	joined, err := meetings.Participant(ctx, room.ID, "u1")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), joined.Joins)

	bob := &model.MeetingParticipant{MeetingID: room.ID, UserID: "u2", UserName: "bob"}
	assert.NoError(t, meetings.Join(ctx, bob))

	joined, err = meetings.Participant(ctx, room.ID, "u1")
	assert.NoError(t, err)
	assert.Equal(t, "ana", joined.UserName)

	people, err := meetings.Participants(ctx, room.ID)
	assert.NoError(t, err)
	assert.Len(t, people, 2)

	present, err := meetings.Present(ctx, room.ID)
	assert.NoError(t, err)
	assert.Equal(t, 2, present)

	gone, err := meetings.Left(ctx, room.ID, "u2", &now)
	assert.NoError(t, err)
	assert.True(t, gone)
	present, err = meetings.Present(ctx, room.ID)
	assert.NoError(t, err)
	assert.Equal(t, 1, present)

	// A return counts: the row revives and joins increments.
	assert.NoError(t, meetings.Join(ctx, bob))
	back, err := meetings.Participant(ctx, room.ID, "u2")
	assert.NoError(t, err)
	assert.Equal(t, int64(2), back.Joins)
	assert.True(t, back.LeftAt == nil)

	late := now.Add(-time.Minute)
	assert.NoError(t, meetings.Seen(ctx, room.ID, "u1", &late))
	stale, err := meetings.Stale(ctx, &now)
	assert.NoError(t, err)
	assert.Len(t, stale, 1)
	assert.Equal(t, "u1", stale[0].UserID)

	line := &model.MeetingLine{ID: model.NewID(), MeetingID: room.ID, UserID: "u1", UserName: "ana", Kind: model.LineSaid, Body: "hello", StartMs: 0, EndMs: 900}
	assert.NoError(t, meetings.AddLine(ctx, line))
	system := &model.MeetingLine{ID: model.NewID(), MeetingID: room.ID, UserID: "u2", UserName: "bob", Kind: model.LineJoin}
	assert.NoError(t, meetings.AddLine(ctx, system))

	lines, err := meetings.Lines(ctx, room.ID, 0)
	assert.NoError(t, err)
	assert.Len(t, lines, 2)
	assert.Equal(t, "hello", lines[0].Body)

	room.State, room.Summary, room.SummaryState = model.MeetingEnded, "## Topics", model.SummaryDone
	assert.NoError(t, meetings.Update(ctx, room))
	got, err = meetings.Get(ctx, room.ID)
	assert.NoError(t, err)
	assert.Equal(t, model.MeetingEnded, got.State)

	assert.NoError(t, meetings.Delete(ctx, room.ID))
	_, err = meetings.Get(ctx, room.ID)
	assert.Equal(t, storage.ErrNotFound, err)
}

// Working returns the ended rooms whose summary a stopped process left running.
func TestMeetingStorage_Working(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	meetings := storage.NewMeetingStorage(newDB(t))
	now := time.Now().UTC().Truncate(time.Millisecond)

	open := &model.Meeting{ID: model.NewID(), State: model.MeetingOpen, SummaryState: model.SummaryWorking, StartedAt: &now, LimitSeconds: model.MeetingLimit}
	assert.NoError(t, meetings.Create(ctx, open))
	done := &model.Meeting{ID: model.NewID(), State: model.MeetingEnded, SummaryState: model.SummaryDone, StartedAt: &now, LimitSeconds: model.MeetingLimit}
	assert.NoError(t, meetings.Create(ctx, done))
	working := &model.Meeting{ID: model.NewID(), State: model.MeetingEnded, SummaryState: model.SummaryWorking, StartedAt: &now, LimitSeconds: model.MeetingLimit}
	assert.NoError(t, meetings.Create(ctx, working))

	rooms, err := meetings.Working(ctx)
	assert.NoError(t, err)
	assert.Len(t, rooms, 1)
	assert.Equal(t, working.ID, rooms[0].ID)
}

// A meeting that is not there answers ErrNotFound.
func TestMeetingStorage_Missing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	meetings := storage.NewMeetingStorage(newDB(t))
	_, err := meetings.Get(ctx, "nope")
	assert.Equal(t, storage.ErrNotFound, err)

	present, err := meetings.Present(ctx, "nope")
	assert.NoError(t, err)
	assert.Equal(t, 0, present)
}

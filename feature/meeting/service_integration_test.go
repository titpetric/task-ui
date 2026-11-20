//go:build integration

package meeting_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/feature/meeting"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/storage"

	_ "modernc.org/sqlite"
)

// newService returns a room service over an in-memory database and a temporary channel root.
func newService(t *testing.T) (*meeting.Service, *model.User, *model.User) {
	t.Helper()
	db, err := storage.OpenAndMigrate(context.Background(), ":memory:", app.Migrations())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	ana := &model.User{ID: model.NewID(), Name: "ana", Role: model.RoleAdmin}
	bob := &model.User{ID: model.NewID(), Name: "bob", Role: model.RoleRunner}
	return &meeting.Service{
		Store:    storage.NewMeetingStorage(db),
		Users:    storage.NewUserStorage(db),
		Channels: &meeting.Channels{Root: t.TempDir()},
		Hub:      meeting.NewHub(),
	}, ana, bob
}

// A room opens, takes participants, logs what they say and ends with its summary state.
func TestService(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	svc, ana, bob := newService(t)

	room, err := svc.Create(ctx, ana, "Standup")
	assert.NoError(t, err)
	assert.Equal(t, "Standup", room.Title)

	view, err := svc.Get(ctx, room.ID)
	assert.NoError(t, err)
	assert.Equal(t, room.ID, view.Meeting.ID)

	rooms, err := svc.List(ctx, 10)
	assert.NoError(t, err)
	assert.Len(t, rooms, 1)

	_, joined, line, err := svc.Join(ctx, ana, room.ID)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), joined.Joins)
	assert.Equal(t, model.LineJoin, line.Kind)

	_, _, again, err := svc.Join(ctx, ana, room.ID)
	assert.NoError(t, err)
	assert.True(t, again == nil)

	_, _, joinedBob, err := svc.Join(ctx, bob, room.ID)
	assert.NoError(t, err)
	assert.Equal(t, model.LineJoin, joinedBob.Kind)

	_, err = svc.Audio(ctx, room, ana, make([]byte, 64))
	assert.NoError(t, err)
	said, err := svc.Line(ctx, room, ana, "hello", 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), said.StartMs)
	assert.Equal(t, int64(2), said.EndMs)

	assert.NoError(t, svc.Heartbeat(ctx, ana, room.ID))

	left, err := svc.Leave(ctx, bob, room.ID)
	assert.NoError(t, err)
	assert.Equal(t, model.LineLeave, left.Kind)

	_, _, back, err := svc.Join(ctx, bob, room.ID)
	assert.NoError(t, err)
	assert.Equal(t, model.LineRejoin, back.Kind)

	_, err = svc.Finish(ctx, bob, room.ID)
	assert.Equal(t, meeting.ErrNotHost, err)

	ended, err := svc.Finish(ctx, ana, room.ID)
	assert.NoError(t, err)
	assert.Equal(t, model.MeetingEnded, ended.State)
	assert.Equal(t, model.SummaryFailed, ended.SummaryState)

	view, err = svc.Get(ctx, room.ID)
	assert.NoError(t, err)
	assert.Equal(t, model.MeetingEnded, view.Meeting.State)

	assert.NoError(t, svc.Delete(ctx, room.ID))
	_, err = svc.Get(ctx, room.ID)
	assert.True(t, errors.Is(err, storage.ErrNotFound))
}

// The sweeper marks a silent participant gone, and ends a room left empty or past its limit.
func TestServiceSweep(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	svc, ana, _ := newService(t)
	now := time.Now().UTC().Truncate(time.Millisecond)
	svc.Now = func() time.Time { return now }

	room, err := svc.Create(ctx, ana, "")
	assert.NoError(t, err)
	assert.Equal(t, "Meeting by ana", room.Title)
	_, _, _, err = svc.Join(ctx, ana, room.ID)
	assert.NoError(t, err)

	// A sighting older than StaleAfter makes the participant gone, and the room empty.
	old := now.Add(-2 * meeting.StaleAfter)
	assert.NoError(t, svc.Store.Seen(ctx, room.ID, ana.ID, &old))
	view, err := svc.Get(ctx, room.ID)
	assert.NoError(t, err)
	assert.True(t, view.Meeting.EmptySince == nil)
	svc.Sweep(ctx)
	view, err = svc.Get(ctx, room.ID)
	assert.NoError(t, err)
	assert.True(t, view.Meeting.EmptySince != nil)
	assert.Equal(t, model.MeetingOpen, view.Meeting.State)

	// Past the empty grace, the room ends.
	empty := now.Add(-2 * meeting.EmptyAfter)
	view.Meeting.EmptySince = &empty
	assert.NoError(t, svc.Store.Update(ctx, view.Meeting))
	svc.Sweep(ctx)
	view, err = svc.Get(ctx, room.ID)
	assert.NoError(t, err)
	assert.Equal(t, model.MeetingEnded, view.Meeting.State)

	// A room past its hour ends on the next sweep.
	late := now.Add(-2 * time.Hour)
	room2, err := svc.Create(ctx, ana, "Late")
	assert.NoError(t, err)
	room2.StartedAt = &late
	assert.NoError(t, svc.Store.Update(ctx, room2))
	svc.Sweep(ctx)
	room2, err = svc.Store.Get(ctx, room2.ID)
	assert.NoError(t, err)
	assert.Equal(t, model.MeetingEnded, room2.State)

	// Start runs the same sweep in the background and stops with the context.
	ctx2, cancel := context.WithCancel(ctx)
	svc.Start(ctx2)
	cancel()
}

// A connected pipe keeps its participant present however quiet the browser becomes.
func TestServiceSweepPipe(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	svc, ana, _ := newService(t)
	now := time.Now().UTC().Truncate(time.Millisecond)
	svc.Now = func() time.Time { return now }

	room, err := svc.Create(ctx, ana, "Quiet")
	assert.NoError(t, err)
	_, _, _, err = svc.Join(ctx, ana, room.ID)
	assert.NoError(t, err)

	pipe := &meeting.Pipe{MeetingID: room.ID, UserID: ana.ID, UserName: ana.Name, Frames: make(chan meeting.Frame, meeting.Buffer)}
	svc.Hub.Join(pipe)
	defer svc.Hub.Leave(pipe)

	old := now.Add(-2 * meeting.StaleAfter)
	assert.NoError(t, svc.Store.Seen(ctx, room.ID, ana.ID, &old))
	svc.Sweep(ctx)

	view, err := svc.Get(ctx, room.ID)
	assert.NoError(t, err)
	assert.True(t, view.Meeting.EmptySince == nil)
	present, err := svc.Store.Present(ctx, room.ID)
	assert.NoError(t, err)
	assert.Equal(t, 1, present)
	lines, err := svc.Store.Lines(ctx, room.ID, 0)
	assert.NoError(t, err)
	for _, line := range lines {
		assert.True(t, line.Kind != model.LineLeave)
	}
}

// A summary a stopped process left working runs again when the service starts.
func TestServiceSummaryResume(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	svc, ana, _ := newService(t)
	assert.NoError(t, svc.Users.Create(ctx, ana))
	svc.Env = func(ctx context.Context, user, role string) (map[string]string, error) {
		return map[string]string{"CLAUDE_CODE_OAUTH_TOKEN": "test"}, nil
	}
	asked := make(chan string, 1)
	svc.Summarize = func(ctx context.Context, env map[string]string, system, prompt, document string) (string, error) {
		asked <- document
		return "## Topics discussed\n- shipping", nil
	}

	room, err := svc.Create(ctx, ana, "Resume")
	assert.NoError(t, err)
	_, _, _, err = svc.Join(ctx, ana, room.ID)
	assert.NoError(t, err)
	_, err = svc.Line(ctx, room, ana, "we ship today", 1)
	assert.NoError(t, err)
	room.State, room.SummaryState = model.MeetingEnded, model.SummaryWorking
	assert.NoError(t, svc.Store.Update(ctx, room))

	pipe := &meeting.Pipe{MeetingID: room.ID, Frames: make(chan meeting.Frame, meeting.Buffer)}
	svc.Hub.Join(pipe)
	defer svc.Hub.Leave(pipe)

	run, cancel := context.WithCancel(ctx)
	defer cancel()
	svc.Start(run)

	document := <-asked
	assert.Contains(t, document, "ana: we ship today")
	frame := <-pipe.Frames
	assert.Contains(t, string(frame.Text), `"type":"summary"`)

	stored, err := svc.Store.Get(ctx, room.ID)
	assert.NoError(t, err)
	assert.Equal(t, model.SummaryDone, stored.SummaryState)
	assert.Contains(t, stored.Summary, "Topics discussed")
}

// A room whose transcript holds no speech fails its summary instead of asking Claude.
func TestServiceSummaryNoSpeech(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	svc, ana, _ := newService(t)
	assert.NoError(t, svc.Users.Create(ctx, ana))
	svc.Env = func(ctx context.Context, user, role string) (map[string]string, error) {
		return map[string]string{"CLAUDE_CODE_OAUTH_TOKEN": "test"}, nil
	}
	asked := make(chan struct{}, 1)
	svc.Summarize = func(ctx context.Context, env map[string]string, system, prompt, document string) (string, error) {
		asked <- struct{}{}
		return "", nil
	}

	room, err := svc.Create(ctx, ana, "Quiet")
	assert.NoError(t, err)
	_, _, _, err = svc.Join(ctx, ana, room.ID)
	assert.NoError(t, err)
	room.State, room.SummaryState = model.MeetingEnded, model.SummaryWorking
	assert.NoError(t, svc.Store.Update(ctx, room))

	pipe := &meeting.Pipe{MeetingID: room.ID, Frames: make(chan meeting.Frame, meeting.Buffer)}
	svc.Hub.Join(pipe)
	defer svc.Hub.Leave(pipe)

	run, cancel := context.WithCancel(ctx)
	defer cancel()
	svc.Start(run)

	<-pipe.Frames
	stored, err := svc.Store.Get(ctx, room.ID)
	assert.NoError(t, err)
	assert.Equal(t, model.SummaryFailed, stored.SummaryState)
	assert.Contains(t, stored.SummaryMessage, "No speech")

	select {
	case <-asked:
		t.Fatal("Claude was asked to summarise an empty transcript")
	default:
	}
}

// An agent is named on an open room by its creator or an admin, and cleared the same way.
func TestService_SetAgent(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	svc, ana, bob := newService(t)
	svc.Agent = &meeting.AgentHooks{
		Valid: func(ctx context.Context, name string) (*model.User, error) {
			if name != "claude" {
				return nil, meeting.ErrAgent
			}
			return &model.User{ID: "agent1", Name: "claude"}, nil
		},
	}

	room, err := svc.Create(ctx, ana, "Agent")
	assert.NoError(t, err)
	_, err = svc.SetAgent(ctx, bob, room.ID, "claude")
	assert.Equal(t, meeting.ErrNotHost, err)

	named, err := svc.SetAgent(ctx, ana, room.ID, "claude")
	assert.NoError(t, err)
	assert.Equal(t, "claude", named.Agent)
	assert.Equal(t, "agent1", named.AgentID)

	_, err = svc.SetAgent(ctx, ana, room.ID, "nobody")
	assert.Equal(t, meeting.ErrAgent, err)

	cleared, err := svc.SetAgent(ctx, ana, room.ID, "")
	assert.NoError(t, err)
	assert.Equal(t, "", cleared.Agent)
	assert.Equal(t, "", cleared.AgentID)
}

// A summary is retried on an ended room alone, and a room ends on demand.
func TestServiceRetry(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	svc, ana, _ := newService(t)

	_, err := svc.Create(ctx, ana, strings.Repeat("a", meeting.TitleLimit+1))
	assert.Equal(t, meeting.ErrTitle, err)

	room, err := svc.Create(ctx, ana, "Retry")
	assert.NoError(t, err)
	_, err = svc.RetrySummary(ctx, ana, room.ID)
	assert.Equal(t, meeting.ErrOpen, err)

	ended, err := svc.Finish(ctx, ana, room.ID)
	assert.NoError(t, err)
	assert.Equal(t, model.SummaryFailed, ended.SummaryState)
	retried, err := svc.RetrySummary(ctx, ana, room.ID)
	assert.NoError(t, err)
	assert.Equal(t, model.SummaryFailed, retried.SummaryState)
	assert.Contains(t, retried.SummaryMessage, "Claude is not set up")

	room2, err := svc.Create(ctx, ana, "Demand")
	assert.NoError(t, err)
	assert.NoError(t, svc.End(ctx, room2, "The host ended it"))
	stored, err := svc.Store.Get(ctx, room2.ID)
	assert.NoError(t, err)
	assert.Equal(t, model.MeetingEnded, stored.State)
}

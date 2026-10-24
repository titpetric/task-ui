package scheduler_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
	"github.com/titpetric/task-ui/scheduler"
)

// board is the Triager of the tests: it records the triages asked of
// it and answers each with its verdict, or its error, once hold, when
// set, closes.
type board struct {
	mu      sync.Mutex
	runs    []string
	users   []string
	verdict string
	err     error
	hold    chan struct{}
}

func (b *board) TriageRun(_ context.Context, user *model.User, runID string) (*model.Note, string, error) {
	b.mu.Lock()
	b.runs = append(b.runs, runID)
	b.users = append(b.users, user.Name)
	hold := b.hold
	b.mu.Unlock()
	if hold != nil {
		<-hold
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.err != nil {
		return nil, "", b.err
	}
	return &model.Note{Body: "noted"}, b.verdict, nil
}

// The board of the notes is the loop's Triager.
func TestTriager(t *testing.T) {
	t.Parallel()
	var triager scheduler.Triager = (*note.Board)(nil)
	assert.NotNil(t, &triager)
	triager = &board{}
	assert.NotNil(t, triager)
}

func (b *board) triaged() []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]string{}, b.runs...)
}

// A failed run of a schedule with triage on gets one triage, as the
// schedule's user; a run that ended well, a schedule without triage, a
// host's or Claude's run, and a loop without a board triage nothing.
func TestLoop_Triage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	b := &board{verdict: note.TriageResolved}
	loop := scheduler.NewLoop(nil, nil, catalogue{}, nil)
	loop.Board = b
	ops := &model.User{ID: "u1", Name: "ops"}
	on := &model.Schedule{ID: "s1", Folder: "infra", Job: "plan", Triage: 1}
	failed := &model.Run{ID: "r1", Folder: "infra", Job: "plan", State: string(model.StateFailed)}
	now := time.Now()
	loop.Now = func() time.Time { return now }

	verdict, err := loop.Triage(ctx, on, ops, failed)
	assert.NoError(t, err)
	assert.Equal(t, note.TriageResolved, verdict)
	assert.Equal(t, []string{"r1"}, b.triaged())
	assert.Equal(t, []string{"ops"}, b.users)

	verdict, err = loop.Triage(ctx, on, ops, failed)
	assert.NoError(t, err)
	assert.Equal(t, "", verdict, "once a run")
	assert.Equal(t, []string{"r1"}, b.triaged())

	killed := &model.Run{ID: "r2", Folder: "infra", Job: "plan", State: string(model.StateKilled)}
	verdict, err = loop.Triage(ctx, on, ops, killed)
	assert.NoError(t, err)
	assert.Equal(t, "", verdict, "the schedule cools down for an hour")
	assert.Equal(t, []string{"r1"}, b.triaged())
	other := &model.Schedule{ID: "s2", Folder: "infra", Job: "plan", Triage: 1}
	verdict, err = loop.Triage(ctx, other, ops, &model.Run{ID: "r2b", Folder: "infra", Job: "plan", State: string(model.StateFailed)})
	assert.NoError(t, err)
	assert.Equal(t, note.TriageResolved, verdict, "another schedule of the job has its own cooldown")
	assert.Equal(t, []string{"r1", "r2b"}, b.triaged())

	now = now.Add(scheduler.TriageCooldown + time.Second)
	verdict, err = loop.Triage(ctx, on, ops, killed)
	assert.NoError(t, err)
	assert.Equal(t, note.TriageResolved, verdict, "a killed run is a failure, triaged once the cooldown passed")
	assert.Equal(t, []string{"r1", "r2b", "r2"}, b.triaged())
	now = now.Add(scheduler.TriageCooldown + time.Second)

	left := []struct {
		name     string
		schedule *model.Schedule
		run      *model.Run
	}{
		{"ok", on, &model.Run{ID: "r3", Folder: "infra", Job: "plan", State: string(model.StateOK)}},
		{"stopped", on, &model.Run{ID: "r4", Folder: "infra", Job: "plan", State: string(model.StateStopped)}},
		{"triage off", &model.Schedule{Folder: "infra", Job: "plan"}, &model.Run{ID: "r5", Folder: "infra", Job: "plan", State: string(model.StateFailed)}},
		{"a host", &model.Schedule{Folder: "_", Job: "@host:web", Triage: 1}, &model.Run{ID: "r6", Folder: "_", Job: "@host:web", Target: "web", State: string(model.StateFailed)}},
		{"Claude's own", on, &model.Run{ID: "r7", Folder: "infra", Job: model.ClaudeJob, State: string(model.StateFailed)}},
		{"no run", on, nil},
	}
	for _, c := range left {
		verdict, err := loop.Triage(ctx, c.schedule, ops, c.run)
		assert.NoError(t, err, c.name)
		assert.Equal(t, "", verdict, c.name)
	}
	assert.Equal(t, []string{"r1", "r2b", "r2"}, b.triaged())

	off := scheduler.NewLoop(nil, nil, catalogue{}, nil)
	verdict, err = off.Triage(ctx, on, ops, failed)
	assert.NoError(t, err)
	assert.Equal(t, "", verdict, "no board")
}

// A board that refuses the triage, for a user without prompt/edit or a
// run no prompt reaches, ends it with a log line and no error; a stop
// ends it with the stop's error.
func TestLoop_TriageRefused(t *testing.T) {
	t.Parallel()
	ops := &model.User{ID: "u1", Name: "ops"}
	on := &model.Schedule{ID: "s1", Folder: "infra", Job: "plan", Triage: 1}
	failed := &model.Run{ID: "r1", Folder: "infra", Job: "plan", State: string(model.StateFailed)}

	for _, refusal := range []error{note.ErrForbidden, note.ErrNoPrompt, note.ErrPromptFailed} {
		b := &board{err: refusal}
		loop := scheduler.NewLoop(nil, nil, catalogue{}, nil)
		loop.Board = b
		verdict, err := loop.Triage(context.Background(), on, ops, failed)
		assert.NoError(t, err, refusal.Error())
		assert.Equal(t, "", verdict)
		assert.Equal(t, []string{"r1"}, b.triaged(), "asked, and refused")
	}

	stopped, cancel := context.WithCancel(context.Background())
	cancel()
	b := &board{err: context.Canceled}
	loop := scheduler.NewLoop(nil, nil, catalogue{}, nil)
	loop.Board = b
	_, err := loop.Triage(stopped, on, ops, failed)
	assert.ErrorIs(t, err, context.Canceled)
}

// While a schedule's triage still works, the next failed run of the
// schedule gets none; a loop's Cooldown shortens the hour.
func TestLoop_TriageOneAtATime(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	b := &board{verdict: note.TriageNoAction, hold: make(chan struct{})}
	loop := scheduler.NewLoop(nil, nil, catalogue{}, nil)
	loop.Board = b
	loop.Cooldown = time.Minute
	now := time.Now()
	loop.Now = func() time.Time { return now }
	ops := &model.User{ID: "u1", Name: "ops"}
	on := &model.Schedule{ID: "s1", Folder: "infra", Job: "plan", Triage: 1}

	done := make(chan string)
	go func() {
		verdict, _ := loop.Triage(ctx, on, ops, &model.Run{ID: "r1", Folder: "infra", Job: "plan", State: string(model.StateFailed)})
		done <- verdict
	}()
	assert.Eventually(t, func() bool { return len(b.triaged()) == 1 }, 5*time.Second, 10*time.Millisecond)
	verdict, err := loop.Triage(ctx, on, ops, &model.Run{ID: "r2", Folder: "infra", Job: "plan", State: string(model.StateFailed)})
	assert.NoError(t, err)
	assert.Equal(t, "", verdict, "the first triage still works")
	close(b.hold)
	assert.Equal(t, note.TriageNoAction, <-done)
	assert.Equal(t, []string{"r1"}, b.triaged())

	verdict, err = loop.Triage(ctx, on, ops, &model.Run{ID: "r3", Folder: "infra", Job: "plan", State: string(model.StateFailed)})
	assert.NoError(t, err)
	assert.Equal(t, "", verdict, "within the cooldown")
	now = now.Add(2 * time.Minute)
	verdict, err = loop.Triage(ctx, on, ops, &model.Run{ID: "r4", Folder: "infra", Job: "plan", State: string(model.StateFailed)})
	assert.NoError(t, err)
	assert.Equal(t, note.TriageNoAction, verdict)
	assert.Equal(t, []string{"r1", "r4"}, b.triaged())
}
